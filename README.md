# zarc

Claude Code + tmux session launcher.

Gerenciador de sessões tmux com TUI interativo para Claude Code, com configuração automatizada do ambiente de desenvolvimento.

## Instalação

```bash
brew tap zarc-tech/tools
brew install zarc
zarc setup
```

## Uso

```bash
# Abrir o TUI (criar/gerenciar sessões)
zarc

# Configurar o ambiente (tmux, CLAUDE.md, alias)
zarc setup
```

## O que o `zarc setup` configura

1. **Verifica dependências** — tmux, Claude Code, git
2. **Configura tmux** — cria `~/.tmux/zarc.conf` com resurrect + continuum
3. **Instala tpm** — gerenciador de plugins do tmux
4. **Configura CLAUDE.md** — adiciona memória persistente por projeto
5. **Configura alias** — detecta fish/zsh/bash automaticamente

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
