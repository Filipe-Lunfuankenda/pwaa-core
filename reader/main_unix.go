//go:build linux || darwin

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
	"runtime"
	"strings"

	"github.com/webview/webview_go"
)

var globalTempDir string

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

	// Registo forçado de mime types comuns
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".mp4", "video/mp4")

	// Handler Customizado
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reqPath := req.URL.Path
		if reqPath == "/" {
			reqPath = "/index.html"
		}
		
		zipPath := strings.TrimPrefix(reqPath, "/")

		var file *zip.File
		for _, f := range r.File {
			if f.Name == zipPath {
				file = f
				break
			}
		}

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

		data, err := io.ReadAll(rc)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		ctype := mime.TypeByExtension(path.Ext(file.Name))
		if ctype == "" {
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
		fmt.Printf("PWAA_URL:%s\n", serverUrl)
		select {}
	}

	// 4. Lançar o navegador Nativo Unix
	tempDir, err := os.MkdirTemp("", "pwaa_session_*")
	if err == nil {
		globalTempDir = tempDir
		defer os.RemoveAll(tempDir)
	}

	w := webview.New(false)
	if w == nil {
		log.Fatal("O motor grafico (WebKit2GTK/Cocoa) nao foi encontrado neste sistema.")
	}
	defer w.Destroy()

	w.SetTitle(filepath.Base(pwaaPath) + " - Leitor PWAA")
	w.SetSize(1024, 768, webview.HintNone)

	// Injeção de JS para intercetar links externos e window.open no Unix
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

	// Binding nativo seguro para abrir no browser padrão do SO (Linux/macOS)
	w.Bind("openExternal", func(targetUrl string) {
		lower := strings.ToLower(targetUrl)
		if strings.HasPrefix(lower, "file:") || strings.HasPrefix(lower, "cmd:") {
			log.Println("Aviso de Segurança: Bloqueada tentativa de executar comando perigoso:", targetUrl)
			return
		}
		
		var cmd *exec.Cmd
		if runtime.GOOS == "darwin" {
			cmd = exec.Command("open", targetUrl)
		} else {
			cmd = exec.Command("xdg-open", targetUrl)
		}
		cmd.Start()
	})

	w.Navigate(serverUrl)
	w.Run()
}
