#!/bin/bash
set -e

GREEN="\033[32m"
YELLOW="\033[33m"
RESET="\033[0m"
CHECK="${GREEN}✓${RESET}"
WARN="${YELLOW}⚠${RESET}"

step() { printf " ${CHECK} %s\n" "$1"; }
skip() { printf " ${CHECK} %s (já instalado)\n" "$1"; }
warn() { printf " ${WARN} %s\n" "$1"; }

echo ""
echo " Claude Orchestrator — Instalação automática"
echo " ─────────────────────────────────────────────"
echo ""

# 1. Homebrew
if command -v brew &>/dev/null; then
  skip "Homebrew"
else
  echo " Instalando Homebrew..."
  /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
  step "Homebrew instalado"
fi

# 2. Go
if command -v go &>/dev/null; then
  skip "Go ($(go version | awk '{print $3}'))"
else
  brew install go
  step "Go instalado"
fi

# 3. tmux
if command -v tmux &>/dev/null; then
  skip "tmux"
else
  brew install tmux
  step "tmux instalado"
fi

# 4. Node
if command -v node &>/dev/null; then
  skip "Node ($(node --version))"
else
  brew install node
  step "Node instalado"
fi

# 5. Claude Code
if command -v claude &>/dev/null; then
  skip "Claude Code"
else
  npm install -g @anthropic-ai/claude-code
  step "Claude Code instalado"
fi

# 6. Git SSH para repos privados
if git config --global --get url."git@github.com:".insteadOf &>/dev/null; then
  skip "Git SSH (repos privados)"
else
  git config --global url."git@github.com:".insteadOf "https://github.com/"
  step "Git configurado para repos privados (SSH)"
fi

# 7. PATH — detectar GOBIN real e adicionar ao PATH
GOBIN=$(go env GOBIN 2>/dev/null)
if [ -z "$GOBIN" ]; then
  GOBIN="$(go env GOPATH 2>/dev/null)/bin"
fi
if [ -z "$GOBIN" ] || [ "$GOBIN" = "/bin" ]; then
  GOBIN="$HOME/go/bin"
fi

SHELL_NAME=$(basename "$SHELL")

add_to_path_rc() {
  local rc_file="$1"
  local gobin_path="$2"
  if [ -f "$rc_file" ] && grep -q "$gobin_path" "$rc_file"; then
    skip "PATH go/bin ($SHELL_NAME)"
  else
    echo "" >> "$rc_file"
    echo "# go bin — Claude Orchestrator" >> "$rc_file"
    echo "export PATH=\"$gobin_path:\$PATH\"" >> "$rc_file"
    step "PATH atualizado ($SHELL_NAME — $rc_file)"
  fi
}

case "$SHELL_NAME" in
  fish)
    if fish -c 'echo $PATH' 2>/dev/null | grep -q "$GOBIN"; then
      skip "PATH go/bin (fish)"
    else
      fish -c "set -Ua fish_user_paths $GOBIN" 2>/dev/null
      step "PATH atualizado (fish)"
    fi
    ;;
  zsh)
    add_to_path_rc "$HOME/.zshrc" "$GOBIN"
    ;;
  bash)
    add_to_path_rc "$HOME/.bashrc" "$GOBIN"
    ;;
  *)
    warn "Shell '$SHELL_NAME' não reconhecido — adicione $GOBIN ao PATH manualmente"
    ;;
esac

# Garantir que GOBIN está no PATH desta sessão
export PATH="$GOBIN:$PATH"

# 8. Instalar claude-orchestrator
echo ""
echo " Instalando claude-orchestrator..."
go install github.com/gildembergleite/claude-orchestrator/cmd/claude-orchestrator@latest
step "claude-orchestrator instalado"

# 9. Localizar o binário instalado e rodar setup
CO_BIN=$(command -v claude-orchestrator 2>/dev/null || echo "$GOBIN/claude-orchestrator")
if [ ! -f "$CO_BIN" ]; then
  # Fallback: procurar onde go install colocou
  CO_BIN=$(find "$(go env GOPATH)" -name claude-orchestrator -type f 2>/dev/null | head -1)
fi

if [ -n "$CO_BIN" ] && [ -f "$CO_BIN" ]; then
  echo ""
  echo " Executando claude-orchestrator setup..."
  echo ""
  "$CO_BIN" setup --no-alias
else
  warn "Binário claude-orchestrator não encontrado — rode 'claude-orchestrator setup' manualmente após reiniciar o terminal"
fi

echo ""
echo " ─────────────────────────────────────────────"
echo " Instalação concluída!"
echo ""
echo " Reinicie o terminal e execute: claude-orchestrator"
echo ""
