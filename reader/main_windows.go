package main

import (
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"unsafe"

	"github.com/jchv/go-webview2"
)

var globalTempDir string

func init() {
	// Garante que o Windows não aplica "zoom de compatibilidade", resolvendo a pixelação/borrado
	user32 := syscall.NewLazyDLL("user32.dll")
	setProcessDPIAware := user32.NewProc("SetProcessDPIAware")
	setProcessDPIAware.Call()
}

func main() {
	if len(os.Args) < 2 {
		log.Println("Uso: pwaareader <ficheiro.pwaa>")
		os.Args = append(os.Args, "../app_teste.pwaa")
	}

	pwaaPath := os.Args[1]

	// 1. Ler o ficheiro ZIP
	r, err := zip.OpenReader(pwaaPath)
	if err != nil {
		log.Fatal("Erro ao abrir o ficheiro PWAA: ", err)
	}
	defer r.Close()

	// 2. Iniciar um servidor HTTP invisível
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Erro ao criar o socket: ", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	serverUrl := fmt.Sprintf("http://127.0.0.1:%d", port)

	// Registo forçado de mime types comuns (para contornar problemas no Windows)
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".mp4", "video/mp4")

	// Handler Customizado: Lê para a RAM e suporta SPA
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reqPath := req.URL.Path
		if reqPath == "/" {
			reqPath = "/index.html"
		}
		
		// Remover a barra inicial para bater certo com os caminhos no ZIP
		zipPath := strings.TrimPrefix(reqPath, "/")

		// Procurar o ficheiro no ZIP
		var file *zip.File
		for _, f := range r.File {
			if f.Name == zipPath {
				file = f
				break
			}
		}

		// Fallback SPA: Se não encontrar, tenta devolver o index.html (React Router, etc)
		if file == nil {
			for _, f := range r.File {
				if f.Name == "index.html" {
					file = f
					break
				}
			}
		}

		if file == nil {
			http.NotFound(w, req)
			return
		}

		rc, err := file.Open()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer rc.Close()

		// LER PARA A MEMÓRIA RAM (Isto resolve o erro 416 dos Vídeos!)
		data, err := io.ReadAll(rc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// ServeContent suporta Range Requests (Seek) nativamente com bytes.Reader
		ctype := mime.TypeByExtension(path.Ext(file.Name))
		if ctype == "" {
			// Detetar tipo se não souber
			ctype = http.DetectContentType(data)
		}
		w.Header().Set("Content-Type", ctype)

		http.ServeContent(w, req, file.Name, file.Modified, bytes.NewReader(data))
	})

	// Rota de Shutdown Gracioso (Anti-Leak)
	http.HandleFunc("/api/shutdown", func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Shutting down..."))
		
		log.Println("A receber ordem de shutdown do Studio...")
		// Limpeza manual imediata do lixo antes do suicídio do processo
		if globalTempDir != "" {
			os.RemoveAll(globalTempDir)
		}
		os.Exit(0)
	})

	go func() {
		log.Printf("PWAA a correr internamente em %s\n", serverUrl)
		if err := http.Serve(listener, nil); err != nil {
			log.Fatal("Erro no micro-servidor: ", err)
		}
	}()

	// 3. Verificação do Modo Headless
	isHeadless := false
	for _, arg := range os.Args {
		if arg == "--headless" {
			isHeadless = true
			break
		}
	}

	if isHeadless {
		// Modo Invisível: Apenas imprime o URL de forma limpa (para o Studio poder ler o stdout)
		fmt.Printf("PWAA_URL:%s\n", serverUrl)
		
		// Mantém o processo vivo a escutar pedidos para sempre, ou até o Studio o matar
		select {}
	}

	// 4. Lançar o navegador Nativo Levíssimo (Se não for Headless)
	// Criar uma pasta temporária para garantir que não sobram resíduos, caches ou cookies (Memory/Disk Leak Prevention)
	tempDir, err := os.MkdirTemp("", "pwaa_session_*")
	if err != nil {
		log.Fatal("Erro ao criar sandbox temporária: ", err)
	}
	globalTempDir = tempDir
	
	// Forçar a Variável de Ambiente para corrigir o bug do WebView2 no Windows que ignora o DataPath
	os.Setenv("WEBVIEW2_USER_DATA_FOLDER", tempDir)
	
	// O defer assegura que quando a janela for fechada, todo o lixo é apagado instantaneamente
	defer os.RemoveAll(tempDir)

	w := webview2.NewWithOptions(webview2.WebViewOptions{
		Debug:          false, // Desativado em produção por segurança
		AutoFocus:      true,
		DataPath:       tempDir,
		WindowOptions: webview2.WindowOptions{
			Title:  filepath.Base(pwaaPath) + " - Leitor PWAA",
			Width:  1024,
			Height: 768,
			Center: true,
		},
	})
	if w == nil {
		messageBoxW := syscall.NewLazyDLL("user32.dll").NewProc("MessageBoxW")
		title, _ := syscall.UTF16PtrFromString("Erro Critico - PWAA Reader")
		msg, _ := syscall.UTF16PtrFromString("O motor grafico (WebView2/Chromium) nao foi encontrado neste sistema.\n\nPara reproduzir ficheiros PWAA, instale o WebView2 Runtime gratuito da Microsoft.\n\nO seu navegador principal sera aberto para o descarregar.")
		messageBoxW.Call(0, uintptr(unsafe.Pointer(msg)), uintptr(unsafe.Pointer(title)), 0x10) // 0x10 = MB_ICONERROR
		exec.Command("rundll32", "url.dll,FileProtocolHandler", "https://developer.microsoft.com/en-us/microsoft-edge/webview2/").Start()
		os.Exit(1)
	}
	defer w.Destroy()

	// Injeção de JS para intercetar links externos e window.open
	w.Init(`
		window.open = function(url) {
			if (url) {
				window.openExternal(url);
			}
			return null;
		};
		window.addEventListener('click', function(e) {
			var a = e.target.closest('a');
			if (a && a.href) {
				var url = a.href;
				var isExternal = (url.startsWith('http') && !url.startsWith('` + serverUrl + `')) || 
				                 (!url.startsWith('http') && !url.startsWith('file') && !url.startsWith('about') && !url.startsWith('javascript'));
				if (isExternal) {
					e.preventDefault();
					window.openExternal(url);
				} else if (a.target === '_blank') {
					e.preventDefault();
					window.location.href = url;
				}
			}
		}, true);
	`)

	// Binding nativo seguro para abrir no browser por defeito
	w.Bind("openExternal", func(targetUrl string) {
		// VERIFICAÇÃO DE SEGURANÇA (Anti-Malware/Anti-RCE)
		lower := strings.ToLower(targetUrl)
		if strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, "cmd:") || strings.HasPrefix(lower, "powershell:") {
			log.Println("Aviso de Segurança: Bloqueada tentativa de executar comando perigoso:", targetUrl)
			return
		}
		cmd := exec.Command("rundll32", "url.dll,FileProtocolHandler", targetUrl)
		cmd.Start()
	})

	w.Navigate(serverUrl)
	w.Run()
}
