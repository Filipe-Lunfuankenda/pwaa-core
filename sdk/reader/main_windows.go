package main

import (
	"archive/zip"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strings"
	"syscall"
	"time"
	"unsafe"
	_ "embed"

	"github.com/jchv/go-webview2"
)

var globalTempDir string
var tempFiles []string
var sessionAuthorized bool = false

func debugLog(msg string) {
	tempDir := os.TempDir()
	logFile := filepath.Join(tempDir, "pwaa_reader_debug.log")
	f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err == nil {
		defer f.Close()
		f.WriteString(time.Now().Format("15:04:05.000") + " - " + msg + "\n")
	}
}

func init() {
	// Garante que o Windows não aplica "zoom de compatibilidade", resolvendo a pixelação/borrado
	user32 := syscall.NewLazyDLL("user32.dll")
	setProcessDPIAware := user32.NewProc("SetProcessDPIAware")
	setProcessDPIAware.Call()
}

func main() {
	os.Setenv("WEBVIEW2_ADDITIONAL_BROWSER_ARGUMENTS", "--disable-web-security")
	if len(os.Args) < 2 {
		log.Println("Uso: pwaareader <ficheiro.pwaa>")
		os.Args = append(os.Args, "../app_teste.pwaa")
	}

	pwaaPath := os.Args[1]

	var injectScriptPath string
	for i := 2; i < len(os.Args)-1; i++ {
		if os.Args[i] == "--inject" {
			injectScriptPath = os.Args[i+1]
			break
		}
	}
	
	var customInjectBytes []byte
	if injectScriptPath != "" {
		b, err := os.ReadFile(injectScriptPath)
		if err == nil {
			customInjectBytes = b
		} else {
			debugLog("Aviso: Falha ao carregar script de injecao: " + err.Error())
		}
	}

	// 1. Ler o ficheiro ZIP
	r, err := zip.OpenReader(pwaaPath)
	if err != nil {
		log.Fatal("Erro ao abrir o ficheiro PWAA: ", err)
	}
	defer r.Close()

	// 2. Iniciar um servidor HTTP invisível com Token de Sessão (Anti-CSRF Local)
	tokenBytes := make([]byte, 16)
	rand.Read(tokenBytes)
	sessionToken := "t_" + hex.EncodeToString(tokenBytes)

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		log.Fatal("Erro ao criar o socket: ", err)
	}

	port := listener.Addr().(*net.TCPAddr).Port
	serverUrl := fmt.Sprintf("http://127.0.0.1:%d/?token=%s", port, sessionToken)
	
	// PRO DRM: Ler Metadados
	var meta struct {
		Password   string `json:"password"`
		ExpiryDate string `json:"expiry_date"`
	}
	for _, f := range r.File {
		if f.Name == "pwaa.meta.json" {
			rc, err := f.Open()
			if err == nil {
				json.NewDecoder(rc).Decode(&meta)
				rc.Close()
			}
			break
		}
	}

	// Registo forçado de mime types comuns (para contornar problemas no Windows)
	mime.AddExtensionType(".js", "application/javascript")
	mime.AddExtensionType(".css", "text/css")
	mime.AddExtensionType(".svg", "image/svg+xml")
	mime.AddExtensionType(".mp4", "video/mp4")
	mime.AddExtensionType(".woff", "font/woff")
	mime.AddExtensionType(".woff2", "font/woff2")
	mime.AddExtensionType(".ttf", "font/ttf")
	mime.AddExtensionType(".eot", "application/vnd.ms-fontobject")
	
	// Suporte para WebAssembly e Modelos 3D
	mime.AddExtensionType(".wasm", "application/wasm")
	mime.AddExtensionType(".glb", "model/gltf-binary")
	mime.AddExtensionType(".gltf", "model/gltf+json")
	mime.AddExtensionType(".obj", "text/plain")

	// Handler Global: Aceita URLs limpos ao nível da raiz (indispensável para React/Next.js/SPA)
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		reqPath := req.URL.Path
		debugLog(fmt.Sprintf("REQ: %s (Referer: %s, Sec-Fetch-Site: %s)", reqPath, req.Header.Get("Referer"), req.Header.Get("Sec-Fetch-Site")))
		cookieName := fmt.Sprintf("pwaa_session_%d", port)
		validToken := false
		
		// 1. Validar via Query Parameter (Primeiro Carregamento)
		qToken := req.URL.Query().Get("token")
		if qToken == sessionToken {
			validToken = true
			sessionAuthorized = true // Sessão autorizada para a vida deste processo
			// Ainda tentamos gravar o Cookie como redundância para navegadores autónomos
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    sessionToken,
				Path:     "/",
				HttpOnly: true,
			})
		} else if sessionAuthorized {
			// 2. Autorização pré-existente (Bypass total para SPA/Redirects/Refreshes)
			validToken = true
		} else {
			// 3. Validar via Sec-Fetch-Site (Nativo do Browser contra Port Scanning / Cross-Site)
			fetchSite := req.Header.Get("Sec-Fetch-Site")
			if fetchSite == "same-origin" || fetchSite == "same-site" {
				validToken = true
			} else {
				// 4. Validar via Referer (Se o host de origem for o nosso próprio microservidor)
				referer := req.Header.Get("Referer")
				if referer != "" {
					refUrl, err := url.Parse(referer)
					if err == nil && refUrl.Host == fmt.Sprintf("127.0.0.1:%d", port) {
						validToken = true
					}
				}
			}
			
			// 5. Redundância: Tentar validar o Cookie caso o Sec-Fetch e o Referer falhem
			if !validToken {
				cookie, err := req.Cookie(cookieName)
				if err == nil && cookie.Value == sessionToken {
					validToken = true
				}
			}
		}

		if !validToken {
			debugLog(fmt.Sprintf("FORBIDDEN: %s (Headers: Referer=%s, Sec-Fetch-Site=%s)", reqPath, req.Header.Get("Referer"), req.Header.Get("Sec-Fetch-Site")))
			http.Error(w, "Forbidden - Invalid PWAA Session", http.StatusForbidden)
			return
		}

		// Rota Unificada: Shutdown Gracioso
		if reqPath == "/api/shutdown" {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Shutting down..."))
			log.Println("A receber ordem de shutdown do Studio...")
			if globalTempDir != "" {
				os.RemoveAll(globalTempDir)
			}
			for _, tmpFile := range tempFiles {
				os.Remove(tmpFile)
			}
			os.Exit(0)
			return
		}

		// Rota Unificada: Abertura Externa
		if reqPath == "/api/open-external" {
			targetUrl := req.URL.Query().Get("url")
			if targetUrl != "" {
				lower := strings.ToLower(targetUrl)
				if !strings.HasPrefix(lower, "file:") && !strings.HasPrefix(lower, "cmd:") && !strings.HasPrefix(lower, "powershell:") {
					exec.Command("rundll32", "url.dll,FileProtocolHandler", targetUrl).Start()
				}
			}
			w.WriteHeader(http.StatusOK)
			return
		}

		// Rota Unificada: Login (Verificação de Senha)
		if reqPath == "/api/login" {
			if req.Method == "POST" {
				body, _ := io.ReadAll(req.Body)
				if string(body) == meta.Password {
					http.SetCookie(w, &http.Cookie{Name: "pwaa_auth", Value: "granted", Path: "/", MaxAge: 86400})
					w.WriteHeader(http.StatusOK)
					return
				}
			}
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if reqPath == "" {
			reqPath = "/"
		}
		
		// 1. Verificação de Expiração (DRM Pro)
		if meta.ExpiryDate != "" {
			exp, err := time.Parse("2006-01-02", meta.ExpiryDate)
			if err == nil && time.Now().After(exp) {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusForbidden)
				w.Write([]byte(`<!DOCTYPE html><html><body style="background:#111;color:#ff4444;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;"><div><h1>⚠️ Acesso Negado</h1><p>Este ficheiro PWAA atingiu a sua data de expiração e foi bloqueado por segurança.</p></div></body></html>`))
				return
			}
		}
		
		// 2. Verificação de Palavra-Passe (DRM Pro)
		if meta.Password != "" && reqPath != "/api/login" {
			cookie, err := req.Cookie("pwaa_auth")
			if err != nil || cookie.Value != "granted" {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Write([]byte(`<!DOCTYPE html><html><body style="background:#111;color:#fff;font-family:sans-serif;display:flex;align-items:center;justify-content:center;height:100vh;flex-direction:column;">
					<h2>🔒 Arquivo Protegido</h2>
					<p>Insira a palavra-passe para aceder a este ficheiro PWAA.</p>
					<form id="lf"><input type="password" id="pw" style="padding:10px;border-radius:5px;border:none;" autofocus><button style="padding:10px 20px;background:#3b82f6;color:white;border:none;border-radius:5px;margin-left:5px;">Desbloquear</button></form>
					<p id="err" style="color:#ff4444;display:none;">Palavra-passe incorreta.</p>
					<script>
					document.getElementById('lf').onsubmit = async function(e) {
						e.preventDefault();
						let p = document.getElementById('pw').value;
						let res = await fetch('/api/login', {method:'POST', body: p});
						if (res.ok) window.location.reload();
						else document.getElementById('err').style.display='block';
					}
					</script>
				</body></html>`))
				return
			}
		}
		
		if reqPath == "/" {
			reqPath = "/index.html"
		}
		
		// Remover a barra inicial para bater certo com os caminhos no ZIP
		zipPath := strings.TrimPrefix(reqPath, "/")

		// Procurar o ficheiro no ZIP (Exato)
		var file *zip.File
		for _, f := range r.File {
			if f.Name == zipPath {
				file = f
				break
			}
		}

		// Fallback 1: Pesquisa Case-Insensitive (Resolve problemas de links e Windows vs Linux)
		if file == nil {
			lowerZipPath := strings.ToLower(zipPath)
			for _, f := range r.File {
				if strings.ToLower(f.Name) == lowerZipPath {
					file = f
					break
				}
			}
		}
		
		// Fallback 2: Pesquisa tolerante a Query Strings guardadas em disco pelo scraper (ex: style.css?v=1)
		if file == nil {
			lowerZipPath := strings.ToLower(zipPath)
			for _, f := range r.File {
				if strings.HasPrefix(strings.ToLower(f.Name), lowerZipPath+"?") {
					file = f
					break
				}
			}
		}

		// Fallback 3: Tentar anexar .html ou .htm (útil para links e scripts que usam URLs limpos ou sem a extensão adicionada pelo scraper)
		if file == nil {
			for _, suffix := range []string{".html", ".htm"} {
				testPath := zipPath + suffix
				for _, f := range r.File {
					if strings.EqualFold(f.Name, testPath) {
						file = f
						break
					}
				}
				if file != nil {
					break
				}
			}
		}

		reqExt := path.Ext(reqPath)
		// Rota SPA: caminhos sem extensão que esperam HTML
		isSPARoute := reqExt == "" && strings.Contains(req.Header.Get("Accept"), "text/html")

		// 1. Fallback SPA: Se for uma rota SPA, tenta devolver o index.html na raiz
		if file == nil && isSPARoute {
			for _, f := range r.File {
				if f.Name == "index.html" {
					file = f
					break
				}
			}
		}

		// 2. Fallback index profundo: Se ainda não encontrou e for rota SPA, procura qualquer index.html
		if file == nil && isSPARoute {
			for _, f := range r.File {
				if strings.HasSuffix(strings.ToLower(f.Name), "index.html") {
					file = f
					break
				}
			}
		}

		// 3. Fallback genérico: Se ainda não encontrou e for rota SPA, devolve o primeiro ficheiro .html
		if file == nil && isSPARoute {
			for _, f := range r.File {
				name := strings.ToLower(f.Name)
				if strings.HasSuffix(name, ".html") || strings.HasSuffix(name, ".htm") {
					file = f
					break
				}
			}
		}

		if file == nil {
				debugLog(fmt.Sprintf("404: %s (zipPath: %s)", reqPath, zipPath))
			http.NotFound(w, req)
			return
		}

		// Extração Inteligente de MIME Type (Ignorar query strings)
		cleanPath := strings.Split(reqPath, "?")[0]
		cleanPath = strings.Split(cleanPath, "#")[0]
		ext := path.Ext(cleanPath)

		ctype := mime.TypeByExtension(ext)
		if ctype == "" {
			switch ext {
			case ".css":
				ctype = "text/css"
			case ".js":
				ctype = "application/javascript"
			case ".woff2":
				ctype = "font/woff2"
			case ".woff":
				ctype = "font/woff"
			case ".ttf":
				ctype = "font/ttf"
			case ".svg":
				ctype = "image/svg+xml"
			default:
				ctype = "application/octet-stream"
			}
		}
		
		debugLog(fmt.Sprintf("FOUND: %s -> %s (Mime: %s)", reqPath, file.Name, ctype))
		
		var rs io.ReadSeeker
		var serveData []byte
		var err error

		isHTML := strings.Contains(ctype, "text/html") || ext == ".html" || ext == ".htm"

		if isHTML {
			var rc io.ReadCloser
			rc, err = file.Open()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			serveData, err = io.ReadAll(rc)
			rc.Close()
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			script := []byte(`
<script>
(function() {
	// 1. Intercetar Links Externos e _blank
	var proxyOpen = window.open;
	window.open = function(url, target, features) {
		if (!url) return null;
		var isExternal = (url.startsWith('http') && !url.startsWith(window.location.origin)) || 
						 (!url.startsWith('http') && !url.startsWith('file') && !url.startsWith('about') && !url.startsWith('javascript') && !url.startsWith('blob'));
		if (isExternal) {
			if (window.openExternal) window.openExternal(url);
			else fetch('/api/open-external?url=' + encodeURIComponent(url));
			return null;
		}
		window.location.href = url;
		return null;
	};

	// 2. Intercetar Clicks e matar _blank
	window.addEventListener('click', function(e) {
		var a = e.target.closest('a');
		if (a && a.href) {
			var url = a.href;
			var isExternal = (url.startsWith('http') && !url.startsWith(window.location.origin)) || 
							 (!url.startsWith('http') && !url.startsWith('file') && !url.startsWith('about') && !url.startsWith('javascript') && !url.startsWith('blob'));
			if (isExternal) {
				e.preventDefault();
				e.stopPropagation();
				if (window.openExternal) window.openExternal(url);
				else fetch('/api/open-external?url=' + encodeURIComponent(url));
			} else if (a.target === '_blank') {
				e.preventDefault();
				e.stopPropagation();
				window.location.href = url;
			}
		}
	}, true);

	// 3. Intercetar Forms com target=_blank
	window.addEventListener('submit', function(e) {
		if (e.target && e.target.target === '_blank') {
			e.target.removeAttribute('target');
		}
	}, true);

	// 4. Limpeza DOM contínua (Mata _blank antes que o utilizador clique)
	var observer = new MutationObserver(function() {
		document.querySelectorAll('a[target="_blank"], form[target="_blank"]').forEach(function(el) {
			el.removeAttribute('target');
		});
	});
	window.addEventListener('DOMContentLoaded', function() {
		observer.observe(document.documentElement, { childList: true, subtree: true, attributes: true, attributeFilter: ['target'] });
		document.querySelectorAll('a[target="_blank"], form[target="_blank"]').forEach(function(el) { el.removeAttribute('target'); });
	});
})();
</script>`)

			if len(customInjectBytes) > 0 {
				customTag := append([]byte("<script>\n"), customInjectBytes...)
				customTag = append(customTag, []byte("\n</script>\n")...)
				
				idx := bytes.LastIndex(bytes.ToLower(serveData), []byte("</body>"))
				if idx == -1 {
					idx = bytes.LastIndex(bytes.ToLower(serveData), []byte("</html>"))
				}
				if idx != -1 {
					newData := make([]byte, 0, len(serveData)+len(customTag)+len(script))
					newData = append(newData, serveData[:idx]...)
					newData = append(newData, script...)
					newData = append(newData, customTag...)
					newData = append(newData, serveData[idx:]...)
					serveData = newData
				} else {
					serveData = append(serveData, script...)
					serveData = append(serveData, customTag...)
				}
			} else {
				idx := bytes.LastIndex(bytes.ToLower(serveData), []byte("</body>"))
				if idx == -1 {
					idx = bytes.LastIndex(bytes.ToLower(serveData), []byte("</html>"))
				}
				if idx != -1 {
					newData := make([]byte, 0, len(serveData)+len(script))
					newData = append(newData, serveData[:idx]...)
					newData = append(newData, script...)
					newData = append(newData, serveData[idx:]...)
					serveData = newData
				} else {
					serveData = append(serveData, script...)
				}
			}
			rs = bytes.NewReader(serveData)
			ctype = "text/html; charset=utf-8"
		} else {
			if file.Method == zip.Store {
				pwaaFd, err := os.Open(pwaaPath)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer pwaaFd.Close()
				offset, err := file.DataOffset()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				rs = io.NewSectionReader(pwaaFd, offset, int64(file.UncompressedSize64))
			} else if file.UncompressedSize64 > 50*1024*1024 {
				var rc io.ReadCloser
				rc, err = file.Open()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer rc.Close()
				tmp, err := os.CreateTemp("", "pwaa_stream_*")
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer tmp.Close()
				tempFiles = append(tempFiles, tmp.Name())
				io.Copy(tmp, rc)
				tmp.Seek(0, 0)
				rs = tmp
			} else {
				var rc io.ReadCloser
				rc, err = file.Open()
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer rc.Close()
				var data []byte
				data, err = io.ReadAll(rc)
				if err != nil {
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				rs = bytes.NewReader(data)
			}
		}

		w.Header().Set("Content-Type", ctype)
		http.ServeContent(w, req, file.Name, file.Modified, rs)
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

	w.Init(`
		window.addEventListener('click', function(e) {
			var el = e.target.closest('a');
			if (el && el.href) {
				var isExternal = !el.href.startsWith(window.location.origin) && !el.href.startsWith('/');
				if (isExternal) {
					e.preventDefault();
					if (window.openExternal) window.openExternal(el.href);
				} else if (el.target === '_blank') {
					e.preventDefault();
					window.location.href = el.href;
				}
			}
		}, true);

		var originalOpen = window.open;
		window.open = function(url, name, specs) {
			if (url && url.startsWith('http') && !url.startsWith(window.location.origin)) {
				if (window.openExternal) window.openExternal(url);
				return null;
			}
			window.location.href = url;
			return null;
		};
	`)

	w.Navigate(serverUrl)
	w.Run()
}
