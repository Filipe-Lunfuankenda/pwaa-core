# ESPECIFICAÇÃO OFICIAL E DOCUMENTAÇÃO DO FORMATO PWAA
**Portable Web Application Archive - Padrão Universal de Arquivamento Offline**

---

## ÍNDICE GERAL
1. **[Introdução e Filosofia](#1-introdução-e-filosofia)**
   - [1.1. O Desafio Histórico do Arquivamento Web](#11-o-desafio-histórico-do-arquivamento-web)
   - [1.2. O Paradigma do PWAA](#12-o-paradigma-do-pwaa)
   - [1.3. O Fim da Dependência da Nuvem](#13-o-fim-da-dependência-da-nuvem)
   - [1.4. Objetivos Primários do Formato](#14-objetivos-primários-do-formato)
2. **[Especificações de Estrutura do Ficheiro (.pwaa)](#2-especificações-de-estrutura-do-ficheiro-pwaa)**
   - [2.1. O Contentor Binário (Norma PKZIP)](#21-o-contentor-binário-norma-pkzip)
   - [2.2. Algoritmos de Compressão Admissíveis](#22-algoritmos-de-compressão-admissíveis)
   - [2.3. Estrutura de Diretórios e Regras de Topologia](#23-estrutura-de-diretórios-e-regras-de-topologia)
   - [2.4. Pontos de Montagem (Entrypoints)](#24-pontos-de-montagem-entrypoints)
   - [2.5. Assinaturas MIME e Magic Bytes](#25-assinaturas-mime-e-magic-bytes)
3. **[O Motor de Leitura (PWAA Reader)](#3-o-motor-de-leitura-pwaa-reader)**
   - [3.1. Arquitetura de Virtualização de Memória (VFS)](#31-arquitetura-de-virtualização-de-memória-vfs)
   - [3.2. Servidor HTTP In-Memory](#32-servidor-http-in-memory)
   - [3.3. Algoritmo de "SPA Fallback"](#33-algoritmo-de-spa-fallback-a-magia-do-react-vue-next)
   - [3.4. Resolução de Range Requests (HTTP 416)](#34-resolução-de-range-requests-http-416-para-multimédia)
   - [3.5. Aceleração Gráfica e Alta Resolução](#35-aceleração-gráfica-e-alta-resolução-dpi-awareness)
   - [3.6. Segurança, Sandboxing e Trapping](#36-segurança-sandboxing-e-trapping-de-links-externos)
4. **[O Motor de Construção (PWAA Builder)](#4-o-motor-de-construção-pwaa-builder)**
   - [4.1. Filosofia de Agnosticidade](#41-filosofia-de-agnosticidade)
   - [4.2. O Modo PACK](#42-o-modo-pack-empacotamento-raw)
   - [4.3. O Modo BUILD](#43-o-modo-build-compilação-nativa-multi-framework)
   - [4.4. O Modo SCRAPE](#44-o-modo-scrape-web-crawler-recursivo-inteligente)
5. **[Algoritmos Avançados do Ecossistema](#5-algoritmos-avançados-do-ecossistema)**
   - [5.1. Breadth-First Search no Scraper](#51-breadth-first-search-no-scraper)
   - [5.2. Mutação DOM e Reescrita de Árvore](#52-mutação-dom-e-reescrita-de-árvore-offline-mapping)
   - [5.3. Interceção de Protocolos de Rede](#53-interceção-de-protocolos-de-rede)
6. **[Integrações e Casos de Uso Industriais](#6-integrações-e-casos-de-uso-industriais)**
   - [6.1. E-Learning em Ambientes Isolados](#61-e-learning-em-ambientes-isolados)
   - [6.2. Documentação Técnica Descentralizada](#62-documentação-técnica-descentralizada)
   - [6.3. Relatórios Médicos Privados e Dashboards](#63-relatórios-médicos-privados-e-dashboards-financeiros)
   - [6.4. Distribuição P2P de Projetos](#64-distribuição-p2p-de-projetos)
7. **[Comparativo Técnico Rigoroso](#7-comparativo-técnico-rigoroso)**
   - [7.1. PWAA vs PDF](#71-pwaa-vs-pdf-portable-document-format)
   - [7.2. PWAA vs EPUB](#72-pwaa-vs-epub)
   - [7.3. PWAA vs XPS / CHM](#73-pwaa-vs-xps--chm)
   - [7.4. PWAA vs Executáveis Electron/Tauri](#74-pwaa-vs-executáveis-electrontauri)
8. **[Referência Oficial de Comandos (CLI)](#8-referência-oficial-de-comandos-cli)**
   - [8.1. Parâmetros e Opções do Builder](#81-parâmetros-e-opções-do-builder)
   - [8.2. Variáveis e Comportamento Adicional](#82-variáveis-e-comportamento-adicional)
   - [8.3. Códigos de Saída (Exit Codes)](#83-códigos-de-saída-exit-codes)
9. **[Guia de Resolução de Problemas](#9-guia-de-resolução-de-problemas-troubleshooting)**
   - [9.1. Erros Críticos de Parsing](#91-erros-críticos-de-parsing)
   - [9.2. Problemas de Assinatura Mime-Type](#92-problemas-de-assinatura-mime-type)
   - [9.3. Anomalias de Playback e Buffering](#93-anomalias-de-playback-e-buffering)
10. **[Especificações Binárias (Apêndice para Desenvolvedores)](#10-especificações-binárias-apêndice-para-desenvolvedores)**
    - [10.1. Central Directory File Header](#101-central-directory-file-header)
    - [10.2. Regras de Checksum e Criptografia](#102-regras-de-checksum-e-criptografia)
    - [10.3. Requisitos para Implementação de Leitores](#103-requisitos-para-implementação-de-leitores-de-terceiros)
11. **[Roteiro de Futuro (Roadmap)](#11-roteiro-de-futuro-roadmap)**
12. **[Perguntas Frequentes Exaustivas (Mega-FAQ)](#12-perguntas-frequentes-exaustivas-mega-faq)**

---

## 1. INTRODUÇÃO E FILOSOFIA

### 1.1. O Desafio Histórico do Arquivamento Web
Desde o final dos anos 90, o mundo da computação tenta, sem grande sucesso, encontrar uma forma elegante de guardar uma página de internet. O formato nativo `.html` que os browsers oferecem quando fazemos "Guardar Como" é catastrófico: cria um ficheiro HTML e, ao lado, uma pasta embaraçosa cheia de imagens estilhaçadas. 
A Microsoft tentou resolver isto com o `.mhtml`, condensando tudo, mas os ficheiros eram pesados e o suporte da indústria quase nulo. Anos depois, aplicações massivas como o Electron começaram a empacotar sites dentro de executáveis pesando, no mínimo, centenas de megabytes.

O grande problema era: "Como arquivamos a Web rica de hoje?". Formatos clássicos como o PDF congelam documentos estáticos, destruindo completamente qualquer interatividade, vídeos ou navegação fluida de Single Page Applications (SPA). O EPUB é desenhado estritamente para texto e e-books estruturados, engasgando com JavaScript pesado. O WARC (Web ARChive) é o padrão de ouro para preservação institucional, mas é assustadoramente complexo para o utilizador comum e exige servidores de reprodução dedicados (Replay Servers).

Faltava no mercado um contentor universal: um ficheiro capaz de armazenar toda a complexidade das frameworks modernas (React, Vue, WebGL) e entregar uma experiência interativa, rápida e encapsulada com um simples duplo-clique, suportado pelas tecnologias nativas dos sistemas operativos modernos. É aqui que nasce o formato PWAA.

### 1.2. O Paradigma do PWAA
O **PWAA (Portable Web Application Archive)** surge como a derradeira solução. Ele não é um executável inchado, nem é um formato morto. Ele é um padrão de ficheiro vivo, desenhado com a exata filosofia da Web Moderna, mas perfeitamente operável sem um byte de tráfego de Internet. 
O seu nome traduz aquilo que ele entrega: a capacidade de colocar a complexidade absurda das Single Page Applications, dos vídeos em HD, e do CSS dinâmico num objeto sólido e transportável, como se fosse um PDF interativo de esteróides.

### 1.3. O Fim da Dependência da Nuvem
Atualmente, qualquer dashboard interativo em React ou sistema documental exige um servidor de alojamento (Vercel, AWS, Netlify). O PWAA liberta os engenheiros da Nuvem. O programador escreve o código, empacota num PWAA, envia por email ou pendrive, e o consumidor final tem uma experiência exata e fluida, independente do encerramento de servidores ou quebras de ligação.

### 1.4. Objetivos Primários do Formato
1. **Total Agnosticidade:** Não interessa a origem. Pode ser puro código à moda antiga, pode ser um projeto massivo compilado de várias frameworks, ou pode ser um site online rasgado (scraped).
2. **Leveza Absoluta:** O formato não transporta navegadores dentro de si. Ele tira proveito dos motores já instalados no sistema operativo anfitrião.
3. **Isolamento de Segurança:** Um PWAA não interage com o disco do utilizador, inibindo vírus tradicionais.
4. **Rapidez:** Sem algoritmos de compressão de alta densidade no acesso, o tempo de leitura (Time-To-Interactive) ronda os parcos milissegundos.

---

## 2. ESPECIFICAÇÕES DE ESTRUTURA DO FICHEIRO (.pwaa)

O ficheiro `.pwaa` não usa criptografia proprietária. Para assegurar a sobrevivência a longo prazo do formato, a base assenta nos standards de arquivo mais difundidos do planeta.

### 2.1. O Contentor Binário (Norma PKZIP)
Todo o `.pwaa` é construído em cima da árvore binária da especificação PKZIP.
Isto significa que, numa situação de extrema necessidade ou corrupção do Leitor Oficial, qualquer utilizador com conhecimentos intermédios de informática pode alterar a extensão para `.zip` e extrair o conteúdo vitalício com o Explorador do Windows, WinRAR ou 7-Zip. A liberdade da informação é imperativa.

### 2.2. Algoritmos de Compressão Admissíveis
O PWAA Standard apenas certifica duas bandeiras de compressão (Compression Flags):
- **0x00 (Store):** Nenhuma compressão. Apenas arquivamento lado a lado. É o método **altamente recomendado** para performance máxima na leitura de média pesada.
- **0x08 (Deflate):** Compressão clássica. Admissível para projetos puramente textuais, mas desaconselhada para PWAA contendo grandes blocos de imagens ou vídeos HD, uma vez que a descompressão "em tempo de voo" (On-the-fly) pode gerar atrasos nos frames.

### 2.3. Estrutura de Diretórios e Regras de Topologia
O interior de um arquivo `.pwaa` validado deve refletir uma topologia "pronta-para-servidor" (Ready-to-Serve Topology). O formato não exige subpastas obrigatórias (além da raiz explícita), permitindo que a estrutura imite a *build* das frameworks de origem.

**Topologia Padrão Esperada:**
```text
<Raiz PWAA>
│
├── index.html            <- OBRIGATÓRIO (O Coração)
├── favicon.ico           <- OPCIONAL mas Recomendado
├── styles/               <- OPCIONAL (Pode ter qualquer nome)
│   └── main.css
├── scripts/              <- OPCIONAL
│   └── app.js
└── assets/               <- OPCIONAL
    ├── logos/
    │   └── marca.png
    └── media/
        └── demonstracao.mp4
```

### 2.4. Pontos de Montagem (Entrypoints)
O Leitor PWAA obriga à existência de um `index.html` no nível zero da árvore. Se um arquivo for empacotado tendo a pasta base dentro dele (ex: `dist/index.html` no nível 1), o Leitor rejeitará o `.pwaa`. O conteúdo deve fluir livremente na raiz.

### 2.5. Assinaturas MIME e Magic Bytes
Historicamente, leitores locais sofrem quando o sistema operativo não reconhece ficheiros (ex: o registo do Windows do utilizador não sabe o que é um `.css`).
O padrão PWAA decreta que o servidor de renderização embutido **DEVE ignorar o SO** e forçar as cabeçalhas MIME baseadas em assinaturas duras:
- Extensões `.js` e `.mjs` -> `application/javascript`
- Extensões `.css` -> `text/css`
- Extensões `.svg` -> `image/svg+xml`
- Extensões `.json` -> `application/json`
Esta regra cimenta a robustez; se funcionar num computador, funcionará em todos.

---

## 3. O MOTOR DE LEITURA (PWAA READER)

O executável `pwaareader.exe` não é um browser convencional; ele é um orquestrador de redes virtuais fechadas.

### 3.1. Arquitetura de Virtualização de Memória (VFS)
Uma falha comum em outras tecnologias de embalagem (como algumas variantes do NW.js e CHM antigos) é descarregar (unzip) o ficheiro de forma fantasma para as pastas `C:\Users\...\AppData\Temp`. Isto é perigoso, deixa lixo sistémico e diminui a segurança corporativa.
O Leitor PWAA nunca escreve dados para o disco físico da máquina anfitriã. Ele mapeia os metadados do ZIP e desenha um Sistema de Ficheiros Virtual (Virtual File System) que aponta diretamente para os bytes dentro do ficheiro `.pwaa`.

### 3.2. Servidor HTTP In-Memory
O motor local não tenta enganar o browser utilizando caminhos do tipo `file:///`, que estão restritos pela CORS (Cross-Origin Resource Sharing) e impediriam chamadas `fetch()` internas ou carregamento de módulos ES6.
Em vez disso, ele levanta uma socket TCP segura na camada interna (`localhost`, por norma `127.0.0.1`), descobrindo dinamicamente uma porta efêmera que não esteja a ser usada pelo computador (ex: `63412`). O ficheiro transforma-se assim, literalmente, num site hospedado.

### 3.3. Algoritmo de "SPA Fallback" (A Magia do React, Vue, Next)
Aplicações modernas mudam o URL da barra de endereço para criar historial de navegação (ex: de `/` para `/galeria`). Mas a página `/galeria.html` não existe; é tudo desenhado no ecrã dinamicamente pelo JavaScript principal.
Servidores fracos crasham e dão "404 Not Found" se o utilizador fizer refresh nesta página.
**O Algoritmo PWAA Fallback entra em ação:**
Quando o Leitor não encontra o ficheiro fisicamente na árvore interna, em vez de atirar erro, ele desvia o tráfego impercetivelmente para o `index.html` central. O JavaScript nativo da tua aplicação (o *Router*) assume então o controlo instantaneamente. O resultado é a perfeição de aplicações complexas offline.

### 3.4. Resolução de Range Requests (HTTP 416) para Multimédia
Este foi o "Santo Graal" tecnológico na génese do formato. Sistemas operativos exigem capacidades de *streaming* por blocos (Range Bytes) para reproduzir ficheiros de vídeo grandes (`.mp4`, `.webm`). Mas os componentes de arquivos ZIP não sabem o que é retroceder num pacote de bytes no seu núcleo clássico.
**A Estratégia de RAM-Bursting:** O Leitor reconhece chamadas a ficheiros e procede à transferência ultrarrápida do bloco de dados necessário para um leitor lógico na memória primária (RAM). Esta conversão instantânea garante que o motor web possa pedir "salta para o minuto 3:40 do vídeo", e a RAM devolve esse exato byte, eliminando de vez os catastróficos Códigos HTTP 416 e falhas de reprodução.

### 3.5. Aceleração Gráfica e Alta Resolução (DPI Awareness)
A aparência é crítica. Sem injeções nativas, janelas desenhadas nativamente no Windows podiam herdar escamas turvas (Efeito Blurriness) provenientes do sistema de ampliação compatível (Scaling 125% ou 150%).
O Leitor PWAA acopla-se ao núcleo `user32.dll` do Windows durante a sua inicialização (fase 0) e declara formalmente **SetProcessDPIAware()**.
Esta ação transfere o controlo da grelha de pixéis do Windows para o WebView. Como resultado, as interfaces, vetores SVG, fontes finas e jogos WebGL ganham exata nitidez de retina, idêntica a navegadores topo de gama.

### 3.6. Segurança, Sandboxing e Trapping de Links Externos
Para que um manual técnico seja imersivo, ele não pode sofrer escapes acidentais.
- Um "Script Guarda-Costas" nativo é atrelado globalmente (no evento raiz do DOM).
- Se o `<a href>` tentar apontar para fora do invólucro do localhost, o guarda-costas estilhaça o evento (PreventDefault), sequestra o URL, e reencaminha a ordem para o Explorador do Windows.
- O ecrã do PWAA permanece inalterado, focado no ficheiro, enquanto o sistema abre o Facebook, Twitter ou Google confortavelmente no browser primário do sistema (Edge/Chrome normal).

---

## 4. O MOTOR DE CONSTRUÇÃO (PWAA BUILDER)

O utilitário `pwaa-builder.exe` é a peça de engenharia pesada do formato. A sua missão é abstrair o caos do desenvolvimento web moderno, fornecendo a interface mais simples humanamente possível para gerar os invólucros finais.

### 4.1. Filosofia de Agnosticidade
O Builder adota uma filosofia "Agnóstica Ativa". Ele sabe que existem centenas de linguagens e ferramentas. Ele não impõe a sua própria estrutura; ele descobre qual foi a infraestrutura que usaste, e conforma-se a ela.

### 4.2. O Modo PACK (Empacotamento Raw)
O comando elementar. Usa-se o modo Pack em ambientes puros, onde se escreveu os ficheiros manualmente e eles já estão prontos para produção. A sua única missão é recolher os dados e aplicar o algoritmo de encriptação binária estipulada pela norma PWAA, garantindo que o `index.html` fica central.
- **Caso Perfeito:** Um portefólio académico desenhado apenas com HTML5 e CSS Vanilla.

### 4.3. O Modo BUILD (Compilação Nativa Multi-Framework)
É aqui que brilha a dedução arquitetural. Na era do Node.js e do ASP.NET, ninguém distribui o código fonte (`.ts`, `.jsx`, `.csproj`). 
Quando o construtor recebe a diretiva `build`, executa o seguinte protocolo de guerra cibernética:
1. **Deteção Orgânica:** Inspeciona o diretório raiz para procurar *DNA* do projeto.
2. **Ciclo Node.js (se encontrar package.json):**
   - Instala as dependências ativamente.
   - Dispara a flag de produção `npm run build`.
3. **Ciclo .NET (se encontrar *.csproj):**
   - Imita um servidor empresarial.
   - Corre `dotnet publish -c Release`.
4. **Ciclo de Caça de Artefatos:** O builder não assume que a saída vai para "dist". Ele procura inteligentemente pelos alvos populares: `out`, `build`, `.next/out`, `dist`, e os complexos `bin/Release/netX.X/publish/wwwroot`.
5. **Empacotamento Final:** Absorve o sub-diretório escondido e sela-o no `.pwaa`. O teu código dev nunca toca no produto final.

### 4.4. O Modo SCRAPE (Web Crawler Recursivo Inteligente)
O orgulho supremo do arsenal. Alguns projetos dependem umbilicalmente de um servidor vivo (ex: um site feito em PHP a extrair artigos de uma base de dados MySQL). É impossível empacotar PHP num sistema sem backend.
A diretiva `scrape` não tenta empacotar o servidor; ela tenta **imitar a visão de um humano a aceder a cada página**.

---

## 5. ALGORITMOS AVANÇADOS DO ECOSSISTEMA

### 5.1. Breadth-First Search no Scraper
Ao invocar `pwaa-builder scrape https://site.com`, o construtor inicia um mecanismo recursivo assíncrono avançado:
1. Começa no Nó A (página base). Descarrega tudo.
2. Extrai dezenas de ligações que pertencem ao Nó A (A.1, A.2, A.3).
3. Avança horizontalmente (Breadth-First) processando e salvaguardando centenas de nós numa pasta temporária hiper-oculta do sistema operativo.
4. Constrói um gigantesco *Snapshot Estático* (Fotografia) da aplicação no seu apogeu. As bases de dados MySQL já não importam, porque as fatias de dados já estão congeladas em blocos `.html`.

### 5.2. Mutação DOM e Reescrita de Árvore (Offline Mapping)
Descarregar dezenas de ficheiros não serviria de nada se as hiperligações continuassem a apontar para "https://...".
A Biblioteca de Mutação do Scraper entra nas cordas do HTML de cada ficheiro descarregado, destrói e reescreve todos os parâmetros cruciais:
- `<a href="/sobre/nos">` vira `href="sobre/nos.html"`
- `<img src="https://cdnsite.com/img.jpg">` vira `src="/assets/misc/img.jpg"`
Isto traduz as fundações voláteis da nuvem num percurso rígido e cimentado.

### 5.3. Interceção de Protocolos de Rede
Em caso extremo de páginas muito caóticas, o modelo garante que se alguma chamada AJAX se perder, as *handlers* do servidor embutido do Leitor bloqueiam vazamentos maciços (data leaks) que corrompam o funcionamento nativo do utilizador.

---

## 6. INTEGRAÇÕES E CASOS DE USO INDUSTRIAIS

A norma `.pwaa` preenche lacunas há muito procuradas pelas multinacionais, startups e até governos.

### 6.1. E-Learning em Ambientes Isolados
Escolas localizadas em regiões rurais ou países em desenvolvimento onde a infraestrutura de Internet é débil. Com PWAA, o Ministério de Educação pode entregar cursos inteiros (desenhados em React com vídeos enormes e animações tridimensionais), arquivados num pequeno disco. Os alunos abrem e estudam interativamente, ignorando a rede global.

### 6.2. Documentação Técnica Descentralizada
Grandes infraestruturas energéticas (refinarias, barragens). As salas de comando muitas vezes não têm autorização para acesso livre à Internet externa (Acesso Intranet Frio). Como distribuir um manual técnico de 5 mil páginas repleto de simulações interativas feitas em Vue.js aos engenheiros? Um ficheiro `.pwaa`. Um duplo clique, e a documentação roda impecavelmente.

### 6.3. Relatórios Médicos Privados e Dashboards Financeiros
Departamentos hospitalares processam dados de ressonâncias magnéticas em portais Web que nunca deveriam cruzar redes não seguras. Os investigadores podem criar Dashboards dinâmicos em Next.js para visualizar o impacto das estatísticas, exportar a `build` para PWAA, e partilhar a PEN USB com toda a comissão hospitalar de modo offline, violando zero regras de confidencialidade (RGPD/HIPAA).

### 6.4. Distribuição P2P de Projetos
Uma agência digital desenha um Portefólio brutalmente pesado e repleto de animações framemotion para um cliente de alto estatuto. Em vez de enviar o tradicional link (que expira, falha ou carece de domínio), a agência anexa um ".pwaa". O cliente corre o Portefólio Web no avião durante a sua viagem, e fica boquiaberto.

---

## 7. COMPARATIVO TÉCNICO RIGOROSO

Para afirmar superioridade de arquitetura, o `.pwaa` tem de obliterar (ou substituir nichos de) antigos líderes na indústria.

### 7.1. PWAA vs PDF (Portable Document Format)
- **PDF:** Criado nos anos 90 com a missão de focar na perfeição da Tinta. Ele existe para ser imprimido. A sua estrutura de interatividade é nula, e inserções de JavaScript são raras e altamente difíceis. Repensar conteúdo (Responsive Design) em ecrãs de telemóveis é catastrófico, forçando o utilizador a fazer "Pan & Scan".
- **PWAA:** Criado para a era digital fluida. Totalmente responsivo; desenhado para adaptação móvel e web nativa. A interatividade não é um acessório; é a fundação. Os vídeos correm fluidos sem plugins estranhos da Adobe.

### 7.2. PWAA vs EPUB
- **EPUB:** Um standard admirável suportado maioritariamente por XML e XHTML simples. Desenhado para eBooks. Falha redondamente quando a "página do livro" tenta ter animações elaboradas, bibliotecas como o Three.js ou integrações de áudio supercomplexas. 
- **PWAA:** Come no ecrã o EPUB ao pequeno-almoço no quesito multimédia e programação funcional (aplicações). O EPUB está preso à formatação literal; o PWAA acolhe uma web rica inteira.

### 7.3. PWAA vs XPS / CHM
- **CHM:** (Compiled HTML Help) O lendário motor de ajudas do Windows XP. Preso num ecossistema pré-histórico dependente do velhinho Internet Explorer e atolado em problemas de segurança terríveis que forçaram a Microsoft a abandoná-lo. 
- **PWAA:** O sucessor espiritual perfeito do CHM. Baseado em segurança de *Sandboxing* estrita, alimentado pelo Edge WebView2 atualizadíssimo que roda Chromium a 1000 à hora.

### 7.4. PWAA vs Executáveis Electron/Tauri
- **Electron:** Empacota runtime inteiro e um motor Chromium dentro de CADA ficheiro da aplicação final. Resulta numa aplicação levíssima pesando absurdos 250MB, destruindo espaço e requerendo instalações na máquina.
- **PWAA:** O empacotamento PWAA é zero. O Leitor já tem o motor Chromium disponibilizado pelo Sistema (WebView2). O ficheiro é limpo e puro.

---

## 8. REFERÊNCIA OFICIAL DE COMANDOS (CLI)

Os comandos foram desenhados com parcimónia, focando na elegância de parâmetros curtos e universais.

### 8.1. Parâmetros e Opções do Builder

```bash
pwaa-builder <comando_ação> <diretório_ou_url_alvo> -o <ficheiro_saida>
```

**A Lista Definitiva de Comandos:**

1. **`pack`** - O Encriptador Cru.
   > Utilização clássica: `pwaa-builder pack ./pasta_site -o site_final.pwaa`
   > Processo: Validação simples da raiz e empacotamento rápido ZIP Store. Não compila linguagens.

2. **`build`** - A Mente Compiladora Múltipla.
   > Utilização clássica: `pwaa-builder build ./meu-app-complexo -o aplicacao_v2.pwaa`
   > Processo: Reconhece os registos das frameworks (`package.json`, `.csproj`), descobre se deve usar gestores de pacotes ou *toolchains* robustas, executa-os num canal temporário, e localiza magicamente a pasta de escoamento correta.

3. **`scrape`** - A Aranha Web.
   > Utilização clássica: `pwaa-builder scrape https://exemplo.net -o devorador.pwaa`
   > Processo: Lança dezenas de tarefas recursivas assíncronas ao URL pretendido. Navega, traduz ligações na nuvem para elos de disco físico, suprime backends não suportáveis em tempo real (como scripts PHP) gravando as reações de renderização puras, e recolhe as centenas ou milhares de ficheiros numa árvore limpa para enclausuramento no `.pwaa`.

### 8.2. Variáveis e Comportamento Adicional
Ao instanciar qualquer um dos comandos acima com caminhos não fechados por plicas duplas (ex: pastas com espaços no Windows), o builder poderá fragmentar o acesso de leitura. Deve envolver nomes com espaços sempre entre plicas, e.g., `pwaa-builder pack "Pasta do Manuel" -o obra.pwaa`.

### 8.3. Códigos de Saída (Exit Codes)
- **Código 0:** Triunfo absoluto. O arquivo `.pwaa` está higienizado e pronto a divulgar.
- **Código 1:** Colapso ou interrupção vital. Falhas de ausência de entrypoint (falta o *index.html*). Em comandos de build, isto reflete falhas inerentes à *toolchain* externa que pararam a linha de montagem.

---

## 9. GUIA DE RESOLUÇÃO DE PROBLEMAS (TROUBLESHOOTING)

Mesmo num ecossistema hiper-controlado, as anomalias originárias das origens ocorrem e devem ser identificadas no local.

### 9.1. Erros Críticos de Parsing
- **"O ecrã fica em branco e não acontece nada quando clico duas vezes no ficheiro .pwaa"**
  Isto nunca é falha do WebView, mas sim falta de conformidade da estrutura interior com a Norma Oficial (ver Secção 2.3). Se descompactaste o pacote numa hierarquia onde o ficheiro `index.html` está encerrado dentro de outra pasta (ex: `minha_obra/index.html` no nível 1 e não 0), o Leitor reverte-se ao protocolo `NotFound`. É impossível ao Leitor tentar adivinhar em que sub-nível escondeste o ponto principal de ancoragem. Solução: Embala apenas o interior absoluto da pasta-fonte.

### 9.2. Problemas de Assinatura Mime-Type
- **"Cores desvanecidas, ficheiros ES6 não carregam"**
  Um problema já contornado na atualização da especificação. Se o host falhar o fornecimento forçado da chave `application/javascript`, o Browser nativo aplica a severidade Web Padrão e trata todos os JS como módulos desconhecidos bloqueando toda a programação dinâmica. O novo motor PWAA tem tabelas Mime Hardcoded blindadas que neutralizam essa falha de interpretação originária em discos virgens do Windows.

### 9.3. Anomalias de Playback e Buffering
- **"Apenas alguns vídeos correm, ou correm aos saltos em máquinas muito antigas"**
  O algoritmo 416 RAM-Bursting (ver 3.4) tem um custo físico subjacente: Ele deposita o conteúdo integral temporariamente na memória estática de curto-prazo da máquina (RAM). Se descarregares e embutires um vídeo Raw de 4.8 GB não comprimido no teu projeto final, computadores antigos equipados com parcos 4GB de RAM poderão experienciar falhas de alocação de registos, crashando todo o *sandbox*. Solução: O PWAA não é a Netflix. Comprima as suas matrizes multimédia para H.264 Web Optimized, priorizando faixas entre os 20MB a 300MB no máximo garantindo execução estelar.

---

## 10. ESPECIFICAÇÕES BINÁRIAS (APÊNDICE PARA DESENVOLVEDORES)

Para programadores e grandes corporações que pretendam instituir Leitores Nativos em plataformas distantes (ex: O Leitor Oficial iOS, ou Leitor Mac-ARM nativo), é crítico entender a leitura dos bytes do container.

### 10.1. Central Directory File Header
Uma norma vital para o sucesso imediato do carregamento da árvore sem bloquear o tempo do processador é a leitura final do ZIP (no rodapé binário). Ao invés de vasculhar os metadados lineares, um leitor robusto procura o `End of central directory record` assinado por `0x06054b50`. Daí, indexa instantaneamente a tabela global num Mapa RAM super rápido.
Qualquer tentativa de encriptação em pacotes corromperá as tentativas de montar o Servidor VFS In-Memory virtual. Se planeias injetar encriptações massivas num `.pwaa`, fá-lo externamente numa caixa encriptadora superior, nunca violando as chaves locais do Formato PWAA.

### 10.2. Regras de Checksum e Criptografia
Os bits flag de segurança na grelha 0x08 PKZIP devem obrigatória e estritamente ostentar valor ZERO. O formato dita transparência máxima.

### 10.3. Requisitos para Implementação de Leitores de Terceiros
Caso decidas programar o teu próprio "Leitor PWAA" nascido de raiz noutra linguagem, a Certificação impõe 4 testes obrigatórios para manter interoperabilidade de formato:
1. Deve forçar os Mime Types.
2. Deve servir ficheiros virtuais com buffers robustos em RAM para não esmagar a interface Web nativa e suportar *Seek* (Saltos).
3. Deve deter a interceptação de âncoras hiperligadas cruzadas (*Cross-Origin Anchors*). A injeção JS na inicialização do DOM é imperativa; o leitor local NÃO DEVE permitir passeios perdidos para a wild web dentro do ecrã local de isolamento sob pena de fuga grave UX.
4. Fallback SPA reativo.

---

## 11. ROTEIRO DE FUTURO (ROADMAP)

O formato `.pwaa` acaba de inaugurar a idade do oiro do armazenamento off-grid. No horizonte das suas expansões de arquitetura prevêem-se desenvolvimentos como:

- **Assinaturas SSL Offline:** Inserção de certificados PKI na ponta do ficheiro, garantindo que o utilizador aterra num ecrã verde ("Autor Verificado") provando criptograficamente que o relatório financeiro que lhes chegou no `.pwaa` não sofreu alterações pelo caminho na infraestrutura empresarial.
- **Indexação por Motor Full-Text Global:** Varredura imediata que cruza todos os HTML empacotados, dando uma caixa de busca universal (Search Tudo) que devolve as respostas em milissegundos a partir da casca protetora, um pouco semelhante ao antigo e amado CHM, mas para o século XXI.

---

## 12. PERGUNTAS FREQUENTES EXAUSTIVAS (MEGA-FAQ)

**P: Existe um limite numérico imposto para o tamanho de um ficheiro PWAA?**
R: A barreira é inteiramente imposta pela lógica dos contentores ZIP subjacentes. A norma Zip64 empurra este limite hipotético para os incríveis domínios dos Exabytes. No entanto, no uso real e prático, não devias ultrapassar a marca de alguns gigabytes ou dezenas de gigabytes, pois prejudicarias brutalmente a portabilidade do ficheiro a níveis corporativos normais (transportar por nuvens, pens). Mas sim, o formato suportará o que o disco conseguir arquivar.

**P: Os bots e algoritmos da web conseguem aceder aos dados no meu PWAA?**
R: Numa ótica estrita de motores de busca clássicos, não. Um PWAA é um cofre cerrado estático no ponto de vista do acesso global em rede. As rotas internas invisíveis no ficheiro virtualizado não são expostas a aranhas fora do leitor. Para efeitos de privacidade e isolamento é perfeito.

**P: Pode um PWAA enviar informações para servidores da internet ativamente mesmo a correr de modo offline local?**
R: Absolutamente Sim e de modo fulcral! É isso que o separa dos documentos mortos. Se embutires dentro do teu projeto uma chamada API Fetch para uma base de dados distante a notificar telemetria ("Este relatório foi aberto"), o Javascript fará esse pedido se a máquina do utilizador tiver internet fisicamente ligada. As ligações externas correm. O que fica local e eterno é o "Corpo" visual e de scripts das janelas.

**P: O meu código fica escondido no PWAA? Protege a minha Propriedade Intelectual (Código-Fonte)?**
R: Relativamente. O nível de compressão impede a leitura humana direta via editores de texto básicos num só lance, e esconde toda a malha de pastas complexas num belíssimo ícone hermético. Mas, em estrita definição forense, é passível de descompressão, tornando-se acessível a terceiros. Aconselha-se as práticas clássicas da indústria moderna Web (ofuscação brutal e Minification extrema com Uglify antes do Empacotamento Builder) caso o intuito primário seja ocultar lógicas secretas. 

**P: É possível correr um servidor de base de dados local completa dentro das janelas virtuais PWAA?**
R: O paradigma em curso veda essas interações por impossibilidade ontológica e segurança nativa total. Backends baseados em linguagens dinâmicas que processam no servidor e bancos de dados lógicos requerem execuções de Processos no SO em background, algo que contradiz a premissa de um arquivo inócuo de dupla-clique. O comando de Scraping ou Compilação (O Camaleão Universal) já mitiga todas as perdas teóricas e práticas, solidificando o conteúdo em fatias e renderizações definitivas puramente Frontend para arquivamento histórico sem mácula. O arquivamento de "back-end dinâmico fechado e em execução simultânea" exigiria o pesado e desastroso modelo arquitetónico das frameworks do passado, que violam o peso ultra-leve do nosso Leitor PWAA e o seu compromisso de abrir o formato de forma virtualizada num ambiente altamente protegido, inibindo malwares e cavalos-de-troia.

---
> *"O PWAA não é apenas um formato; é a cápsula do tempo perfeita para a era da Internet Moderna. Ele devolve o poder e a posse absoluta da informação ao utilizador final, sem correntes corporativas nem rendas da nuvem."*

*Especificação Oficial PWAA - The Offline Archival Standard*
