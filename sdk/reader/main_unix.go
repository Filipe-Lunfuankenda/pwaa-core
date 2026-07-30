//go:build linux || darwin

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
	"runtime"
	"strings"
	"time"

	"github.com/webview/webview_go"
)

var globalTempDir string
var tempFiles []string
var sessionAuthorized bool = false

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
			log.Println("Aviso: Falha ao carregar script de injecao: " + err.Error())
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

	// Registo forçado de mime types comuns
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
					var cmd string
					var args []string
					switch runtime.GOOS {
					case "darwin":
						cmd = "open"
						args = append(args, targetUrl)
					case "linux":
						cmd = "xdg-open"
						args = append(args, targetUrl)
					}
					if cmd != "" {
						exec.Command(cmd, args...).Start()
					}
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
			http.NotFound(w, req)
			return
		}

		// Extração Inteligente de MIME Type (Ignorar query strings que possam estar presas no nome do ficheiro)
		ext := path.Ext(reqPath)
		if ext == "" {
			ext = path.Ext(file.Name)
		}
		if idx := strings.Index(ext, "?"); idx != -1 {
			ext = ext[:idx]
		}
		ctype := mime.TypeByExtension(ext)
		var rs io.ReadSeeker

		if file.Method == zip.Store {
			// Acesso direto ao disco rígido (Zero RAM)
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
			
			if ctype == "" {
				ctype = "application/octet-stream"
			}
		} else if file.UncompressedSize64 > 50*1024*1024 {
			// Ficheiro comprimido gigante (> 50 MB)
			// Extrair para disco temporário para não explodir a RAM
			rc, err := file.Open()
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
			if ctype == "" {
				ctype = "application/octet-stream"
			}
		} else {
			// Ficheiro pequeno (< 50 MB), ler para a RAM (Máxima Velocidade)
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
			
			if ctype == "" {
				ctype = http.DetectContentType(data)
			}
			
			// Injeção universal de interceção de links
			if strings.Contains(ctype, "text/html") {
				script := []byte(`
<script>
(function() {
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
		window.location.href = url; // Forçar interno na mesma janela
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
				window.location.href = url; // Forçar links internos a abrir na mesma frame
			}
		}
	}, true);
})();
</script>`)

				if len(customInjectBytes) > 0 {
					customTag := append([]byte("<script>\n"), customInjectBytes...)
					customTag = append(customTag, []byte("\n</script>\n")...)
					
					idx := bytes.LastIndex(bytes.ToLower(data), []byte("</body>"))
					if idx == -1 {
						idx = bytes.LastIndex(bytes.ToLower(data), []byte("</html>"))
					}
					if idx != -1 {
						newData := make([]byte, 0, len(data)+len(customTag)+len(script))
						newData = append(newData, data[:idx]...)
						newData = append(newData, script...)
						newData = append(newData, customTag...)
						newData = append(newData, data[idx:]...)
						data = newData
					} else {
						data = append(data, script...)
						data = append(data, customTag...)
					}
				} else {
					idx := bytes.LastIndex(bytes.ToLower(data), []byte("</body>"))
					if idx == -1 {
						idx = bytes.LastIndex(bytes.ToLower(data), []byte("</html>"))
					}
					if idx != -1 {
						newData := make([]byte, 0, len(data)+len(script))
						newData = append(newData, data[:idx]...)
						newData = append(newData, script...)
						newData = append(newData, data[idx:]...)
						data = newData
					} else {
						data = append(data, script...)
					}
				}
			}
			rs = bytes.NewReader(data)
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
