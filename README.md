# Claude Orchestrator

Gerenciamento de persistência de sessões Claude Code.

Gerenciador de sessões tmux com TUI interativo para Claude Code, com configuração automatizada do ambiente de desenvolvimento.

## Instalação (comando único)

**Zsh/Bash:**
```bash
bash <(curl -fsSL https://raw.githubusercontent.com/gildembergleite/claude-orchestrator/main/install.sh)
```

**Fish:**
```fish
curl -fsSL https://raw.githubusercontent.com/gildembergleite/claude-orchestrator/main/install.sh | bash
```

Esse comando instala e configura automaticamente:
- Homebrew, Go, tmux, Node (se ausentes)
- Claude Code
- Git SSH (`insteadOf` HTTPS → SSH para `github.com`)
- PATH do Go
- claude-orchestrator CLI + setup completo do ambiente

Após a instalação, reinicie o terminal e execute:
```bash
claude-orchestrator
```

## Uso

```bash
# Abrir o TUI (criar/gerenciar sessões)
claude-orchestrator

# Reconfigurar o ambiente
claude-orchestrator setup
```

### Navegação no TUI

- **Setas** — navegar na lista
- **Enter** — selecionar/entrar no diretório
- **Delete/Backspace** — voltar ao diretório anterior
- **Esc** — cancelar
- **Digite** — filtrar diretórios por nome

## O que o `claude-orchestrator setup` configura

1. **Verifica dependências** — tmux, Claude Code, git
2. **Configura tmux** — cria `~/.tmux/claude-orchestrator.conf` com resurrect + continuum
3. **Instala tpm** — gerenciador de plugins do tmux
4. **Configura CLAUDE.md** — adiciona memória persistente por projeto
5. **Configura alias** — escolha entre `claude-orchestrator + co`, só `claude-orchestrator`, ou nome personalizado

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
