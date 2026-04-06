# Install Script Design

**Data:** 2026-04-06
**Branch:** dev

## Resumo

Script bash (`install.sh`) que automatiza toda a instalação do zarc em um único comando. O dev roda via `gh` e o script cuida de dependências, configuração de repos privados, PATH e setup do ambiente.

## Comando de instalação

```bash
bash <(gh api repos/zarc-tech/zarc-claude-orchestrator/contents/install.sh --jq '.content' | base64 -d)
```

Pré-requisito: `gh` instalado e autenticado.

## Passos do script (em ordem)

| # | Passo | Comando | Idempotente |
|---|-------|---------|-------------|
| 1 | Verifica Homebrew | Instala se ausente | Sim |
| 2 | Verifica Go | `brew install go` | Sim |
| 3 | Verifica tmux | `brew install tmux` | Sim |
| 4 | Verifica Node | `brew install node` | Sim |
| 5 | Verifica Claude Code | `npm install -g @anthropic-ai/claude-code` | Sim |
| 6 | Configura Git SSH | `git config --global url."git@github.com:".insteadOf "https://github.com/"` | Sim |
| 7 | Configura GOPRIVATE | `go env -w GOPRIVATE="github.com/zarc-tech/*"` | Sim |
| 8 | Adiciona go/bin ao PATH | Detecta fish/zsh/bash, adiciona se ausente | Sim |
| 9 | Instala zarc | `go install github.com/zarc-tech/zarc-claude-orchestrator/cmd/zarc@latest` | Sim |
| 10 | Roda zarc setup | `zarc setup` (tmux, tpm, plugins, CLAUDE.md, alias) | Sim |

## Detalhes de implementação

### Detecção de shell (passo 8)

Mesma lógica de `internal/setup/shell.go`:
- **Fish:** `set -Ux fish_user_paths $HOME/go/bin $fish_user_paths` (se não contém)
- **Zsh:** adiciona `export PATH="$HOME/go/bin:$PATH"` ao `~/.zshrc` (se não contém)
- **Bash:** adiciona `export PATH="$HOME/go/bin:$PATH"` ao `~/.bashrc` (se não contém)

### Output

Cada passo imprime checkmark verde ou warning amarelo, mesmo estilo do `zarc setup`:
```
 ✓ Homebrew instalado
 ✓ Go instalado (go1.25.5)
 ✓ tmux instalado
 ✓ Node instalado
 ✓ Claude Code instalado
 ✓ Git configurado para repos privados
 ✓ GOPRIVATE configurado
 ✓ PATH atualizado (zsh)
 ✓ zarc instalado
 ✓ zarc setup concluído
```

### Restrições

- macOS apenas (Homebrew como package manager)
- Requer `gh` autenticado
- Requer SSH key configurada no GitHub

## Arquivos impactados

| Arquivo | Ação |
|---------|------|
| `install.sh` | Novo |
| `README.md` | Modificar (atualizar instruções de instalação) |

## Fora de escopo

- Suporte a Linux/apt
- Instalação do `gh` CLI
- Geração de SSH keys
