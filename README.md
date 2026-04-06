# Claude Orchestrator

Gerenciamento de persistência de sessões Claude Code.

Gerenciador de sessões tmux com TUI interativo para Claude Code, com configuração automatizada do ambiente de desenvolvimento.

## Pré-requisitos

### 1. Instalar Go

**macOS (Homebrew):**
```bash
brew install go
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt update
sudo apt install -y golang-go
```

**Ou baixe diretamente:** https://go.dev/dl/

Verifique a instalação:
```bash
go version
```

### 2. Configurar o GOPATH no PATH

Certifique-se de que `$GOPATH/bin` está no seu PATH.

**Fish:**
```fish
fish_add_path $HOME/go/bin
```

**Zsh/Bash:** adicione ao `~/.zshrc` ou `~/.bashrc`:
```bash
export PATH="$HOME/go/bin:$PATH"
```

Recarregue o terminal após a alteração.

### 3. Instalar tmux

**macOS:**
```bash
brew install tmux
```

**Linux (Ubuntu/Debian):**
```bash
sudo apt install -y tmux
```

### 4. Instalar Claude Code

```bash
npm install -g @anthropic-ai/claude-code
```

## Instalação

```bash
go install github.com/zarc-tech/zarc-claude-orchestrator/cmd/zarc@latest
```

Verifique a instalação:
```bash
zarc --help
```

## Configuração inicial

Após a instalação, rode o setup para configurar o ambiente:

```bash
zarc setup
```

O setup irá:
1. **Verificar dependências** — tmux, Claude Code, git
2. **Configurar tmux** — cria `~/.tmux/zarc.conf` com resurrect + continuum
3. **Instalar tpm** — gerenciador de plugins do tmux
4. **Configurar CLAUDE.md** — adiciona memória persistente por projeto
5. **Configurar alias** — escolha entre `zarc + claude`, só `zarc`, ou nome personalizado

## Uso

```bash
# Abrir o TUI (criar/gerenciar sessões)
zarc

# Reconfigurar o ambiente
zarc setup
```

### Navegação no TUI

- **Setas** — navegar na lista
- **Enter** — selecionar/entrar no diretório
- **Delete/Backspace** — voltar ao diretório anterior
- **Esc** — cancelar
- **Digite** — filtrar diretórios por nome

## Desenvolvimento

```bash
make build    # compila o binário
make test     # roda os testes
make run      # compila e executa
make lint     # go vet
```

## Release

```bash
git tag v1.0.0
git push origin v1.0.0
# GitHub Actions roda goreleaser automaticamente
```
