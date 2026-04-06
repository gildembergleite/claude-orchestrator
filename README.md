# Claude Orchestrator

Gerenciamento de persistência de sessões Claude Code.

Gerenciador de sessões tmux com TUI interativo para Claude Code, com configuração automatizada do ambiente de desenvolvimento.

## Instalação (comando único)

**Pré-requisito:** [GitHub CLI (`gh`)](https://cli.github.com/) instalado e autenticado, com SSH key configurada no GitHub.

**Zsh/Bash:**
```bash
bash <(gh api repos/zarc-tech/zarc-claude-orchestrator/contents/install.sh --jq '.content' | base64 -d)
```

**Fish:**
```fish
gh api repos/zarc-tech/zarc-claude-orchestrator/contents/install.sh --jq '.content' | base64 -d | bash
```

Esse comando instala e configura automaticamente:
- Homebrew, Go, tmux, Node (se ausentes)
- Claude Code
- Git SSH para repos privados zarc-tech
- GOPRIVATE para Go modules
- PATH do Go
- zarc CLI + setup completo do ambiente

Após a instalação, reinicie o terminal e execute:
```bash
zarc
```

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

## O que o `zarc setup` configura

1. **Verifica dependências** — tmux, Claude Code, git
2. **Configura tmux** — cria `~/.tmux/zarc.conf` com resurrect + continuum
3. **Instala tpm** — gerenciador de plugins do tmux
4. **Configura CLAUDE.md** — adiciona memória persistente por projeto
5. **Configura alias** — escolha entre `zarc + claude`, só `zarc`, ou nome personalizado

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
