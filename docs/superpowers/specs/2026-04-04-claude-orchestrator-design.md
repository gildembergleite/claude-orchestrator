# claude-orchestrator (zarc) — Design Spec

**Data:** 2026-04-04
**Repo:** `zarc-tech/claude-orchestrator`
**Tap:** `zarc-tech/homebrew-tools`
**Comando CLI:** `zarc`

## Objetivo

Transformar o script bash `zarc-claude-tmux-orchestrator.sh` em um CLI Go distribuível via Homebrew, para que a equipe de desenvolvimento instale e configure o ambiente Claude Code + tmux com um único comando.

## Arquitetura do Projeto

```
zarc-tech/claude-orchestrator/
├── cmd/
│   └── zarc/
│       └── main.go              # entrypoint
├── internal/
│   ├── tui/
│   │   ├── app.go               # modelo principal bubbletea
│   │   ├── menu.go              # componente de menu com setas
│   │   ├── input.go             # componente de input (diretório, nome)
│   │   └── banner.go            # banner ASCII "ZARC"
│   ├── tmux/
│   │   ├── client.go            # wrapper para comandos tmux
│   │   └── session.go           # criar, listar, attach, kill sessões
│   ├── claude/
│   │   └── resolver.go          # localizar binário do Claude Code
│   └── setup/
│       ├── setup.go             # orquestrador do zarc setup
│       ├── tmux.go              # criar ~/.tmux/zarc.conf + source no tmux.conf
│       ├── claude.go            # append memória persistente no CLAUDE.md
│       ├── shell.go             # detectar shell e configurar alias
│       └── tpm.go               # instalar tpm + plugins
├── configs/
│   ├── zarc.tmux.conf           # template do tmux config
│   └── claude-memory.md         # template da seção de memória
├── .goreleaser.yml              # config do goreleaser
├── go.mod
├── go.sum
├── Makefile                     # atalhos para dev local
└── README.md
```

## Dependências Go

- `charmbracelet/bubbletea` — framework TUI
- `charmbracelet/lipgloss` — estilização terminal
- `spf13/cobra` — CLI commands (root + setup)

## Comando `zarc` — TUI Principal

### Estados do TUI (bubbletea)

1. **Banner** — exibe o ASCII "ZARC" estilizado com lipgloss
2. **Menu principal** — lista:
   - `[+] Nova sessão`
   - Sessões tmux existentes (listadas dinamicamente)
3. **Nova sessão** → input sequencial:
   - Input de diretório (com expansão de `~`, validação de existência)
   - Input de nome da sessão (default: basename do diretório)
   - Cria sessão tmux + attach
4. **Sessão existente** → submenu:
   - Attach
   - Kill (com confirmação)
   - Voltar

### Comportamento especial

- Se já está dentro do tmux → pula o TUI, lança `claude` direto
- Navegação: setas, j/k, Enter, q para sair

### Resolução do Claude Code

Procura em ordem:
1. `~/.npm-global/bin/claude`
2. `~/.local/bin/claude`
3. `/usr/local/bin/claude`
4. `npm root -g` → `.bin/claude`
5. Fallback: `npx @anthropic-ai/claude-code`

## Comando `zarc setup`

Executa em etapas sequenciais com feedback visual (spinners/checkmarks). Cada etapa é idempotente — rodar várias vezes é seguro.

### Etapa 1 — Verificar dependências

- tmux instalado? Se não → instrui `brew install tmux`
- Claude Code instalado? Se não → instrui como instalar
- Git instalado? (necessário para tpm)

### Etapa 2 — Configurar tmux

- Cria `~/.tmux/zarc.conf` com:
  ```
  # ─── Terminal compatibility ───
  set -g extended-keys on
  set -gs terminal-features 'xterm*:extkeys'

  # ─── Plugins ───
  set -g @plugin 'tmux-plugins/tpm'
  set -g @plugin 'tmux-plugins/tmux-resurrect'
  set -g @plugin 'tmux-plugins/tmux-continuum'

  # ─── Resurrect + Continuum ───
  set -g @continuum-restore 'on'
  set -g @continuum-save-interval '5'
  set -g @resurrect-capture-pane-contents 'on'
  ```
- Verifica se `~/.tmux.conf` já tem `source-file ~/.tmux/zarc.conf`
  - Se não tem → adiciona a linha
  - Se já tem → pula

### Etapa 3 — Instalar tpm + plugins

- Se `~/.tmux/plugins/tpm` não existe → `git clone https://github.com/tmux-plugins/tpm`
- Executa `~/.tmux/plugins/tpm/bin/install_plugins` para instalar resurrect e continuum

### Etapa 4 — Configurar CLAUDE.md global

- Se `~/.claude/CLAUDE.md` não existe → cria com a seção de memória
- Se existe → verifica se já contém "Memória de Sessão"
  - Se não contém → faz append
  - Se já contém → pula
- Conteúdo da seção:
  ```markdown
  ## Memória de Sessão

  - **Ao iniciar qualquer nova sessão**, verificar se existe um arquivo de memória para o diretório/repositório atual no sistema de memória (`~/.claude/projects/<project-path>/memory/`).
  - Se **não existir**, criar imediatamente um arquivo de memória do tipo `project` com o contexto inicial do repositório (nome do projeto, stack, objetivo geral, branch atual, etc.).
  - **A cada iteração relevante na sessão**, atualizar o arquivo de memória do projeto com o que foi feito, decisões tomadas, e contexto importante — garantindo que sessões futuras nunca percam o histórico do que já foi realizado.
  - O objetivo é manter continuidade entre sessões: qualquer nova conversa deve poder retomar de onde a anterior parou, sem que o usuário precise re-explicar o contexto.
  ```

### Etapa 5 — Configurar alias do shell

- Verifica se `zarc` já está no PATH (instalação via Homebrew coloca em `/opt/homebrew/bin/`)
  - Se está no PATH → pula esta etapa com mensagem "zarc já está no PATH"
- Se não está no PATH, detecta o shell ativo: fish, zsh, bash
  - **Fish**: cria função `zarc` em `~/.config/fish/functions/zarc.fish`
  - **Zsh**: adiciona `alias zarc="/caminho/completo/zarc"` em `~/.zshrc`
  - **Bash**: adiciona `alias zarc="/caminho/completo/zarc"` em `~/.bashrc`
- Se o alias/função já existe → pula

### Output esperado

```
 ✓ Dependências verificadas
 ✓ tmux configurado (~/.tmux/zarc.conf)
 ✓ tpm + plugins instalados
 ✓ CLAUDE.md configurado (memória persistente)
 ✓ Alias configurado (fish)

 Pronto! Execute 'zarc' para começar.
```

## Distribuição via Homebrew

### GoReleaser

- Builds: `darwin/arm64`, `darwin/amd64`, `linux/amd64`, `linux/arm64`
- Binary name: `zarc`
- Homebrew tap: `zarc-tech/homebrew-tools`

### Fluxo de release

1. Cria tag `git tag v1.0.0`
2. Push da tag dispara GitHub Action
3. GoReleaser compila, cria release no GitHub, publica formula no tap

### Instalação para a equipe

```bash
brew tap zarc-tech/tools
brew install zarc
zarc setup
```

### GitHub Actions (CI/CD)

- `ci.yml` — roda em push/PR: `go build`, `go test`, `go vet`, `golangci-lint`
- `release.yml` — roda em tags `v*`: goreleaser build + publish

## Fora do Escopo

- Gerenciamento de múltiplos projetos/workspaces
- Integração com git dentro do TUI
- Auto-update do zarc
- Suporte a Windows
