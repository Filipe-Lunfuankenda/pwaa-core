# PWAA SDK (Software Development Kit)

Este SDK (Software Development Kit) contém todas as ferramentas oficiais necessárias para suportar, gerar e renderizar ficheiros `.pwaa` no seu próprio sistema, firmware, dispositivo ou aplicação.

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