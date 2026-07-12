### 🇵🇹 O que é o Formato PWAA? (Português)

O **PWAA (Portable Web Application Archive)** é um padrão universal de arquivamento desenhado para encapsular e distribuir aplicações web completas, manuais e sites num único ficheiro que funciona 100% offline. Ele liberta aplicações complexas da dependência de servidores na nuvem, permitindo que corram nativamente em qualquer máquina com total interatividade.

**Arquitetura do Ficheiro:**

* A nível binário, um ficheiro `.pwaa` é um contentor compatível com a norma PKZIP.


* O formato utiliza compressão Store (0x00) para máxima velocidade de acesso ou Deflate (0x08) para compressão de texto.


* Exige a presença de um ficheiro `index.html` na raiz do diretório interno para ser considerado válido.



**Motor de Leitura e Renderização:**

* **Virtualização In-Memory (Zero-Leak):** O leitor nunca extrai os ficheiros para o disco rígido, mapeando o conteúdo do ZIP diretamente para a memória RAM. Isto garante segurança corporativa e apaga qualquer rasto (cookies, cache) assim que a janela é fechada.


* **Servidor HTTP Local:** Para contornar as severas restrições CORS dos navegadores, o leitor levanta um micro-servidor efêmero (`127.0.0.1`), fazendo com que o código web acredite estar alojado online.


* **SPA Fallback:** Se a aplicação web pedir um ficheiro inexistente (como uma rota virtual do React ou Vue), o servidor devolve instantaneamente os bytes do `index.html`, garantindo que o *routing* moderno funciona perfeitamente offline.


* **RAM-Buffering para Multimédia:** Resolve os erros HTTP 416 (Range Not Satisfiable) alocando vídeos diretamente na RAM, permitindo saltos instantâneos (*scrubbing*) na reprodução de media pesada.


* **Sandboxing de Links:** Um script injetado globalmente interceta hiperligações externas (`https://`) e força a sua abertura no navegador padrão do sistema operativo, mantendo a janela do PWAA isolada e segura.


* **Aceleração HiDPI:** No Windows, o leitor declara-se como `SetProcessDPIAware` diretamente à API do sistema, garantindo nitidez absoluta em ecrãs de alta resolução.



**Motor de Criação (PWAA Builder):**

* **Modo PACK:** Empacota pastas web estáticas tradicionais aplicando as assinaturas binárias do PWAA.


* **Modo BUILD:** Deteta a tecnologia do projeto (Next.js, React, ASP.NET), instala dependências automaticamente, compila o código para produção e arquiva a pasta de saída correta.


* **Modo SCRAPE:** Funciona como um rastreador web (Web Crawler) utilizando o algoritmo Breadth-First Search. Descarrega sites recursivamente, reescreve a árvore DOM (convertendo caminhos absolutos em relativos) e gera um clone offline do portal.

---

## ⚖️ Conformidade Legal e Privacidade (RGPD, CCPA, UK-GDPR)

Como o formato `.pwaa` executa aplicações web num ambiente nativo offline, ele introduz um paradigma único de isolamento e privacidade de dados. Este formato e os respetivos leitores oficiais foram desenhados com a filosofia de **"Privacy by Design"** (Privacidade desde a Conceção), garantindo conformidade natural com as leis de dados globais, nomeadamente o **Regulamento Geral sobre a Proteção de Dados (RGPD)** da União Europeia, as leis do Reino Unido (UK-GDPR) e da América do Norte (ex: CCPA/CPRA).

### 1. Garantia Zero-Leak (Isolamento Total de Dados)
A regra de ouro da arquitetura do SDK exige que qualquer Leitor `.pwaa` isole a sua execução numa **Sandbox Efémera** (RAM ou Pasta Temporária Volátil). Assim que o ficheiro `.pwaa` é encerrado, **todos os cookies, localStorage, IndexedDB, caches e dados de sessão gerados pela visualização são instantaneamente obliterados pelo sistema operativo.**
Nenhum dado comportamental ou estado do utilizador sobrevive ao fecho da janela do leitor.

### 2. Isenção de Consentimento de Cookies
Devido ao mecanismo Zero-Leak, qualquer cookie que uma página `.pwaa` tente instalar (mesmo que estritamente não essencial) é tratado como um "Cookie de Sessão Volátil isolado". Como o Leitor destrói o rasto local após o encerramento, e operando 100% offline, **o armazenamento não interage de forma persistente com o equipamento terminal do utilizador**. Isto atenua drasticamente os requisitos punitivos da Diretiva ePrivacy (Lei dos Cookies da UE).

### 3. Transferência de Dados Transfronteiriça (Offline-First)
Um ficheiro `.pwaa` roda o seu próprio servidor local (`127.0.0.1`). Em páginas estáticas e documentações:
- O processamento é inteiramente local (Edge Computing no próprio aparelho).
- Não há partilha passiva de endereços de IP com CDNs externos ou Analytics a terceiros, exceto quando explicitamente interage com chamadas de rede dinâmicas programadas pelo criador do `.pwaa`.
A retenção e análise de dados enquadra-se na posse física do dispositivo do utilizador, dispensando complexos Modelos de Transferência de Dados (Standard Contractual Clauses) exigidos pela UE para os EUA/UK.

### 4. Responsabilidade do Desenvolvedor (O Criador do Ficheiro)
Embora o ecossistema PWAA garanta segurança local de nível militar contra rastreio cruzado:
- Se empacotar num `.pwaa` chamadas ativas de Telemetria (como Google Analytics, Meta Pixels) que liguem à Internet, continua a ter a responsabilidade de requerer consentimento e providenciar Avisos de Privacidade no seio da interface gráfica, tal como num site normal.

---

### 🇬🇧 What is the PWAA Format? (English)

The **PWAA (Portable Web Application Archive)** is a universal archiving standard designed to encapsulate and distribute complete web applications, manuals, and websites within a single file that operates 100% offline. It frees complex applications from cloud server dependency, allowing them to run natively on any machine while retaining full interactivity.

**File Architecture:**

* At the binary level, a `.pwaa` file is a container fully compatible with the PKZIP standard.


* The format utilizes Store compression (0x00) for maximum access speed or Deflate (0x08) for text compression.


* It strictly requires the presence of an `index.html` file at the root of its internal directory to be considered a valid archive.



**Reading and Rendering Engine:**

* **In-Memory Virtualization (Zero-Leak):** The reader never extracts files to the local hard drive, mapping the ZIP contents directly into RAM instead. This ensures corporate security and wipes all traces (cookies, cache) the exact millisecond the window is closed.


* **Local HTTP Server:** To bypass strict browser CORS restrictions, the reader spins up an ephemeral micro-server (`127.0.0.1`), tricking the web code into behaving as if it were hosted online.


* **SPA Fallback:** If the web application requests a non-existent file (such as a React or Vue virtual route), the server instantly returns the `index.html` bytes, ensuring modern routing works flawlessly offline.


* **RAM-Buffering for Multimedia:** Solves HTTP 416 (Range Not Satisfiable) errors by allocating video files directly into RAM, enabling instant playback scrubbing for heavy media.


* **Link Sandboxing:** A globally injected script intercepts external hyperlinks (`https://`) and forces them to open in the operating system's default browser, keeping the PWAA window strictly isolated and secure.


* **HiDPI Acceleration:** On Windows, the reader explicitly declares itself as `SetProcessDPIAware` directly to the system API, guaranteeing absolute visual sharpness on high-resolution displays.



**Creation Engine (PWAA Builder):**

* **PACK Mode:** Packages traditional static web folders by applying the PWAA binary signatures.


* **BUILD Mode:** Detects the project's underlying technology (Next.js, React, ASP.NET), automatically installs dependencies, compiles the code for production, and archives the correct output folder.


* **SCRAPE Mode:** Acts as a recursive Web Crawler utilizing a Breadth-First Search algorithm. It downloads websites, rewrites the DOM tree (converting absolute paths to relative ones), and generates a fully offline clone of the portal.

---

## ⚖️ Legal Compliance & Privacy (GDPR, CCPA, UK-GDPR)

Because the '.pwaa' format runs web applications in a native offline environment, it introduces a unique paradigm of data isolation and privacy. This format and its official readers have been designed with the philosophy of **"Privacy by Design"** (Privacy by Design)**, ensuring natural compliance with global data laws, namely the **General Data Protection Regulation (GDPR)** of the European Union, the laws of the United Kingdom (UK-GDPR) and North America (e.g. CCPA/CPRA).

### 1. Zero-Leak Guarantee (Total Data Isolation)
The golden rule of SDK architecture requires any .pwaa Reader to isolate its execution in an Ephemeral Sandbox (RAM or Volatile Temp Folder). Once the '.pwaa' file is closed, **all cookies, localStorage, IndexedDB, caches, and session data generated by the view are instantly obliterated by the operating system.**
No behavioral data or user status survives the closing of the reader window.

### 2. Cookie Consent Waiver
Due to the Zero-Leak mechanism, any cookie that a '.pwaa' page attempts to install (even if strictly non-essential) is treated as an "isolated Volatile Session Cookie". Because the Reader destroys the local trail after shutdown, and operates 100% offline, **the storage does not persistently interact with the user's terminal equipment**. This drastically mitigates the punitive requirements of the ePrivacy Directive (EU Cookie Law).

### 3. Cross-Border Data Transfer (Offline-First)
A .pwaa file runs its own local server (127.0.0.1). On static pages and documentation:
- Processing is entirely local (Edge Computing on the device itself).
- There is no passive sharing of IP addresses with external CDNs or third-party analytics, except when explicitly interacting with dynamic network calls programmed by the creator of the '.pwaa'.
Data retention and analysis falls under the physical possession of the user's device, dispensing with complex Data Transfer Models (Standard Contractual Clauses) required by the EU for the US/UK.

### 4. Developer's Responsibility (The File Creator)
While the PWAA ecosystem ensures military-grade local security against cross-tracking:
- If you package in a '.pwaa' active Telemetry calls (such as Google Analytics, Meta Pixels) that connect to the Internet, you are still responsible for requesting consent and providing Privacy Notices within the graphical interface, just as on a normal website.
