# PWAA SDK (Software Development Kit)

Este SDK (Software Development Kit) contém todas as ferramentas oficiais necessárias para suportar, gerar e renderizar ficheiros `.pwaa` no seu próprio sistema, firmware, dispositivo ou aplicação.

## O Que Está Incluído Neste SDK?

1. **`pwaa-builder.exe`** (O Motor de Criação)
   Ferramenta de linha de comandos (CLI) que converte código web, frameworks ou URLs online para o formato PWAA. Pode ser invocado de forma programática pelo seu software.
   - Uso básico: `pwaa-builder pack ./pasta -o ficheiro.pwaa`
   - Scrape recursivo: `pwaa-builder scrape https://exemplo.com -o site.pwaa`
   - Build universal: `pwaa-builder build ./meu-projeto -o app.pwaa`

2. **`pwaareader.exe`** (O Motor de Renderização Oficial)
   Um leitor ultra-leve com aceleração HiDPI e servidor HTTP virtualizado In-Memory. Permite a reprodução instantânea dos ficheiros.

3. **`PWAA_DOCUMENTATION.md`** (O Whitepaper Oficial)
   A especificação de alto nível do PWAA, detalhando toda a engenharia por trás do formato.

---

## Como Criar o Seu Próprio Visualizador PWAA Nativo

Caso não pretenda utilizar o nosso `pwaareader.exe` e queira **embutir o suporte PWAA diretamente no seu próprio programa, browser ou firmware**, o processo é incrivelmente simples, pois o PWAA foi desenhado para ser universal.

Para ler e apresentar um ficheiro `.pwaa` nativamente em qualquer linguagem (C++, Rust, Python, C#), o seu software deve seguir estes **3 passos arquitetónicos**:

### Passo 1: Descompactação Virtual (In-Memory)
Um ficheiro `.pwaa` é estruturalmente um ficheiro ZIP. 
O seu software **não deve** extrair os ficheiros para o disco rígido (por questões de segurança e velocidade). Em vez disso, utilize uma biblioteca de ZIP da sua linguagem para mapear o conteúdo do ficheiro diretamente para a Memória RAM.

### Passo 2: O Canal de Comunicação (Mini-Servidor ou Esquema Customizado)
Como os motores web (WebViews, Chromium, WebKit) bloqueiam scripts em ficheiros locais (`file:///`), tem duas opções para alimentar o HTML ao ecrã:

- **Opção A (A mais fácil - HTTP Localhost):** O seu programa levanta um mini-servidor HTTP local (`127.0.0.1`) numa porta aleatória (ex: 50431). Sempre que o servidor receber um pedido `/foto.png`, o seu código vai à RAM (ao ZIP), lê a `foto.png` e devolve-a ao browser.
- **Opção B (A mais nativa - Custom Protocol):** Se estiver a usar WebViews modernas (como WebView2 em C# ou C++), pode registar um protocolo customizado (ex: `pwaa://`). Sempre que a WebView pedir `pwaa://app/index.html`, o seu código intercepta o pedido nativamente, lê os bytes do ZIP na memória e atira-os para o ecrã.

### Passo 3: Regras de Ouro da Renderização
Para que a sua implementação seja 100% compatível com a norma PWAA, deve assegurar:
1. **SPA Fallback:** Se a WebView pedir um ficheiro que não existe no ZIP (ex: `/contactos`), o seu servidor não deve devolver o Erro 404, mas sim devolver os bytes do `index.html`. Isto permite que frameworks como o React funcionem offline!
2. **Streaming de Vídeo (HTTP 416):** Para suportar vídeos, o seu leitor deve suportar "Range Requests" (saltar blocos de bytes do vídeo). Aloque o ficheiro de vídeo para um "Buffer" na memória RAM antes de o enviar.
3. **Mime Types:** Garanta que a sua resposta HTTP contém as cabeçalhas corretas (ex: `.js` tem de devolver `application/javascript`, caso contrário o browser rejeita-o).
4. **Isolamento de Links:** Injete um script (JavaScript) na WebView que impeça os utilizadores de abrirem links de internet (`https://`) dentro do Leitor PWAA. Links externos devem ser empurrados para o browser padrão do sistema.
5. **Garantia de Zero Fugas (Anti-Leak):** Motores WebView tendem a gravar cookies, localStorage e caches localmente por defeito. Qualquer Leitor PWAA deve, obrigatoriamente, ser instanciado a apontar para uma "Pasta Temporária de Sessão" (na RAM ou Temp) e eliminá-la ativamente no exato milissegundo em que a janela do Leitor for fechada, não deixando nenhum rastro para trás após a visualização.
6. **Graceful Fallback:** Caso a máquina onde está a correr não tenha motores WebView instalados (ex: falta de Chromium/WebKit), o Leitor não deve "crashar" silenciosamente. Deve instanciar um alerta visual informando o utilizador que lhe falta o "WebView Runtime" e idealmente atirar o utilizador para a página oficial de download.

Seguindo estes 3 passos, **qualquer software ou firmware no mundo** pode tornar-se num leitor nativo do formato `.pwaa`!


## 📥 Guia de Instalação e Execução do PWAA Reader

O ecossistema PWAA foi desenhado com foco absoluto na portabilidade. O formato `.pwaa` pode ser reproduzido nativamente em qualquer sistema operativo, sem necessitar de instalar servidores web pesados ou configurar ambientes de desenvolvimento.

## 1. Onde Descarregar (Releases)

Acede à nossa página oficial de **Releases** no GitHub para obteres os binários mais recentes, compilados e otimizados:
👉 **[Descarregar Última Release do PWAA Core](https://github.com/Filipe-Lunfuankenda/pwaa-core/releases)**

Dentro do pacote `.zip` da release, encontrarás as seguintes ferramentas prontas a usar (Zero Instalação):
*   🖥️ **Desktop:** `pwaareader-windows.exe`, `pwaareader-mac` (Intel/Apple Silicon) e `pwaareader-linux`.
*   📱 **Mobile SDK:** Binários de biblioteca `pwaa-mobile-sdk.aar` (Android) e `pwaa-sdk.xcframework` (iOS).

---

## 2. Como Executar (Desktop)

O `pwaareader` é um executável autónomo (Standalone). Não necessita de instalação, basta abrir o terminal na pasta onde o descarregaste e passar-lhe o ficheiro `.pwaa` que desejas ler.

### 🪟 Windows

No Windows, pode arrastar o ficheiro `.pwaa` para o executável, ou utilizar a linha de comandos:

```powershell
# Reproduzir um ficheiro PWAA
.\pwaareader-windows.exe o_meu_projeto.pwaa

# Modo Silencioso (Headless - apenas inicia o servidor e devolve o URL interno)
.\pwaareader-windows.exe o_meu_projeto.pwaa --headless
```

#### A. Para abrir/ler um ficheiro PWAA
```powershell
.\sdk\pwaareader.exe projeto.pwaa
```

#### B. Converter uma pasta Web simples (Modo Pack)
```powershell
.\sdk\pwaa-builder.exe pack .\sites\nanovisio -o nanovisio.pwaa
```

#### C. Converter um projeto complexo (Modo Build)
```powershell
.\sdk\pwaa-builder.exe build .\meu_projeto_react -o app-moderna.pwaa
```

#### D. Clonar um Site da Internet (Modo Scrape)
```powershell
.\sdk\pwaa-builder.exe scrape https://exemplo.com -o copia-online.pwaa
```

---

### 🐧 Linux

No Linux, atribua permissões de execução antes do primeiro uso:

```bash
# Dar permissão de execução
chmod +x pwaareader-linux

# Reproduzir o ficheiro PWAA
./pwaareader-linux o_meu_projeto.pwaa

# Modo Silencioso (Headless)
./pwaareader-linux o_meu_projeto.pwaa --headless
```

---

### 🍏 MacOS

Tal como no Linux, atribua permissão de execução e execute via terminal. O sistema Mac baseia-se no seu motor Cocoa de alta performance nativa.

```bash
# Dar permissão de execução
chmod +x pwaareader-mac

# Reproduzir o ficheiro PWAA
./pwaareader-mac o_meu_projeto.pwaa
```

---

## 3. Como Integrar (Mobile SDK - Android e iOS)

Diferente da versão Desktop que é um programa executável, a versão Mobile é fornecida sob a forma de uma **Biblioteca Nativa (SDK)** para integração direta nas suas próprias aplicações móveis.

### 🤖 Android (Java/Kotlin)

Dentro da Release, descarregue o ficheiro `pwaa-mobile-sdk.aar` e adicione-o à pasta `app/libs` do seu projeto Android Studio.

```kotlin
// Importar a classe gerada pelo SDK GoMobile
import pwaa.mobile.sdk.Sdk

// Inicializar e arrancar o Servidor Interno PWAA
val engineStatus = Sdk.initSDK()
val fileInfo = Sdk.readFile("/caminho/absoluto/no/telemovel/app.pwaa")

// Posteriormente, carregue o localhost interno numa WebView Android
```

### 🍎 iOS (Swift/Objective-C)

Descarregue a framework `pwaa-sdk.xcframework` da Release e arraste-a para o seu projeto no Xcode (garanta que está marcada como *Embed & Sign*).

```swift
// Importar o módulo
import Pwaa_mobile_sdk

// Iniciar a máquina PWAA e ler um ficheiro local do dispositivo
let status = Pwaa_mobile_sdkInitSDK()
let readerData = Pwaa_mobile_sdkReadFile("/caminho/no/ios/app.pwaa")

// Usar uma WKWebView para apresentar o conteúdo local servido pela framework
```

---
*(PWAA Padrão Oficial - Construído para a era Offline)*