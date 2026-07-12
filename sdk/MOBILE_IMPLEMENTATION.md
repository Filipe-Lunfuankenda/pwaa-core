# Guia de Implementação Mobile PWAA (Android & iOS)

A filosofia da Web Móvel assenta em Sistemas Operativos fortemente vigiados pela Google (Android) e Apple (iOS). Nestes ambientes, não existem executáveis genéricos de terminal (como `.exe` ou binários Linux). Todo o software deve ser distribuído sob a forma de Aplicações com interface visual (pacotes APK/AAB no Android, e IPA no iOS).

Portanto, o suporte PWAA nos telemóveis **não se atinge atirando um executável para o SDK**, mas sim fornecendo a "Planta Arquitetónica" aos programadores mobile. Quando uma equipa Mobile importar os conceitos abaixo para os seus projetos nativos no Xcode ou Android Studio, a sua App adquire imediatamente o "poder" de ler ficheiros `.pwaa`.

---

## 1. Arquitetura Android (Kotlin / Java)

No Android, a arquitetura é extremamente amigável graças à poderosa `android.webkit.WebView`.

**A Mecânica Passo a Passo:**
1. **O Servidor Embutido:** O programador importa uma biblioteca HTTP leve, como o **NanoHTTPD** (ou Ktor). 
2. **Descompactação Virtual:** O servidor HTTP recebe o caminho do ficheiro `.pwaa` selecionado pelo utilizador. Utilizando a classe nativa `java.util.zip.ZipFile`, o servidor lê os bytes sob demanda (On-the-Fly) e mapeia os Mime-Types.
3. **Range Requests:** O programador estende o `NanoHTTPD` para responder ao código `206 Partial Content`. Isto é vital porque os motores de vídeo Android (ExoPlayer/HTML5) usam extensivamente chamadas fragmentadas para ler `.mp4`.
4. **WebView Invocação:** O programador cria uma Atividade (Activity) contendo a WebView.
   - Ativa o Javascript: `webView.settings.javaScriptEnabled = true`
   - Isola o armazenamento (Garantia Zero Fugas): A WebView nativa permite definir a CacheMode e limpar os dados locais `webView.clearCache(true)`.
   - Navega para: `webView.loadUrl("http://localhost:8080/index.html")`

**Interceção de Links Externos no Android:**
Basta fazer um "Override" ao `WebViewClient`:
```kotlin
webView.webViewClient = object : WebViewClient() {
    override fun shouldOverrideUrlLoading(view: WebView, request: WebResourceRequest): Boolean {
        val url = request.url.toString()
        if (url.startsWith("http") && !url.startsWith("http://localhost")) {
            // Empurrar o link para o Google Chrome do utilizador
            val intent = Intent(Intent.ACTION_VIEW, Uri.parse(url))
            startActivity(intent)
            return true // Bloqueia a navegação dentro do PWAA
        }
        return false // Deixa continuar localmente
    }
}
```

---

## 2. Arquitetura iOS e iPadOS (Swift / Objective-C)

No universo Apple, a classe `WKWebView` (WebKit) é um dos motores mais blindados e rápidos do planeta. A Apple proíbe por segurança o uso de ficheiros locais diretos para recursos complexos, logo a arquitetura do "Servidor Local" é obrigatória.

**A Mecânica Passo a Passo:**
1. **O Servidor Embutido:** A framework de eleição no iOS é o **GCDWebServer** (ou Swifter). 
2. **Descompactação Virtual:** Utiliza-se a biblioteca nativa ou o popular `ZIPFoundation` para ler o `.pwaa` armazenado nos Documentos do iPhone.
3. **Mime Types:** A Apple é impiedosa com os MIME Types. Um ficheiro `.js` enviado com o Mime errado resultará num ecrã totalmente branco (Strict MIME Checking). O servidor GCDWebServer deve forçar respostas precisas.
4. **WKWebView Invocação:**
   - Para apagar rastros e manter Zero Fugas no fecho, deve-se usar um Pool de dados efémero: `WKWebsiteDataStore.nonPersistent()`
   - Navega-se chamando a URL gerada pelo servidor invisível.

**Interceção de Links Externos no iOS:**
Recorre-se ao protocolo `WKNavigationDelegate`:
```swift
func webView(_ webView: WKWebView, decidePolicyFor navigationAction: WKNavigationAction, decisionHandler: @escaping (WKNavigationActionPolicy) -> Void) {
    if let url = navigationAction.request.url {
        if url.absoluteString.hasPrefix("http") && !url.absoluteString.hasPrefix("http://localhost") {
            // Abrir no Safari
            UIApplication.shared.open(url)
            decisionHandler(.cancel) // Trava o fluxo interno
            return
        }
    }
    decisionHandler(.allow)
}
```

---
**Conclusão Mobile:** 
O formato PWAA brilha nos telemóveis. Permite que um utilizador de iPhone descarregue um portefólio Web pesando 200MB, e o corra enquanto vai de metro sem rede, com interações a 60fps (sem lag de Internet). Com estas plantas arquitetónicas, qualquer programador Mobile de nível médio recria o "PWAA Reader" no seu SO num par de dias de trabalho.
