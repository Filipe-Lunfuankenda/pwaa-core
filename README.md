### 🇵🇹 O que é o Formato PWAA? (Português)

O **PWAA (Portable Web Application Archive)** é um padrão universal de arquivamento desenhado para encapsular e distribuir aplicações web completas, manuais e sites num único ficheiro que funciona 100% offline. Ele liberta aplicações complexas da dependência de servidores na nuvem, permitindo que corram nativamente em qualquer máquina com total interatividade.

**Arquitetura do Ficheiro:**

* A nível binário, um ficheiro `.pwaa` é um contentor compatível com a norma PKZIP.


* O formato utiliza compressão Store (0x00) para máxima velocidade de acesso ou Deflate (0x08) para compressão de texto.


* Exige obrigatoriamente a presença de um ficheiro `index.html` na raiz do diretório interno para ser considerado válido.



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
