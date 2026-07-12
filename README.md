
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

### O mistério do "Comando não reconhecido"
O terminal do Windows (PowerShell) tem uma medida de segurança muito chata: ele recusa-se a correr executáveis apenas pelo nome se eles não estiverem instalados nas entranhas do sistema (ao contrário dos comandos normais como `cd` ou `git`).

Para dizeres ao Windows *"Quero correr o programa que está exatamente aqui na minha frente"*, tens obrigatoriamente de colocar um ponto e uma barra (**`.\`**) antes do nome!

Aqui tens a tua "Cábula" definitiva com os comandos exatos que deves usar na raiz do teu projeto. Copia e cola-os para ver a magia a acontecer:

#### 1. Para abrir/ler um ficheiro PWAA
Se tiveres um ficheiro chamado `projeto.pwaa` na tua pasta, abres o Leitor assim:
```powershell
.\sdk\pwaareader.exe projeto.pwaa
```

#### 2. Converter uma pasta Web simples (Modo Pack)
Pega numa pasta onde programaste HTML/CSS puro (por exemplo a pasta `sites/nanovisio`) e empacota-a de forma instantânea:
```powershell
.\sdk\pwaa-builder.exe pack .\sites\nanovisio -o nanovisio.pwaa
```

#### 3. Converter um projeto complexo (Modo Build - React/Vue/Node)
Se estiveres numa pasta com ficheiros cruéis cheios de dependências (`package.json`), o builder instala tudo, compila para código limpo de produção e arquiva:
```powershell
.\sdk\pwaa-builder.exe build .\A_TUA_PASTA_REACT -o app-moderna.pwaa
```

#### 4. Clonar um Site da Internet (Modo Scrape)
Para sacares todo um site que está na nuvem e imortalizá-lo offline num formato `.pwaa`:
```powershell
.\sdk\pwaa-builder.exe scrape https://exemplo.com -o copia-online.pwaa
```

A única diferença entre os comandos dos manuais e o teu terminal foi o facto de o Windows exigir que declares o caminho rigoroso do EXE (ou seja `.\sdk\pwaa-builder.exe`). Se quiseres, no futuro podemos adicionar o SDK às Variáveis de Ambiente do teu computador, e aí sim poderás escrever apenas `pwaa-builder` a partir de qualquer pasta do Windows!

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

*(PWAA Padrão Oficial - Construído para a era Offline)*