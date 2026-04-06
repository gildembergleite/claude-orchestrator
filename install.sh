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

# 7. GOPRIVATE
CURRENT_GOPRIVATE=$(go env GOPRIVATE 2>/dev/null || echo "")
if echo "$CURRENT_GOPRIVATE" | grep -q "github.com/zarc-tech"; then
  skip "GOPRIVATE"
else
  go env -w GOPRIVATE="github.com/zarc-tech/*"
  step "GOPRIVATE configurado"
fi

# 8. PATH — detectar GOBIN real e adicionar ao PATH
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

# 9. Instalar zarc
echo ""
echo " Instalando zarc..."
go install github.com/zarc-tech/zarc-claude-orchestrator/cmd/zarc@latest
step "zarc instalado"

# 10. Localizar o binário instalado e rodar setup
ZARC_BIN=$(command -v zarc 2>/dev/null || echo "$GOBIN/zarc")
if [ ! -f "$ZARC_BIN" ]; then
  # Fallback: procurar onde go install colocou
  ZARC_BIN=$(find "$(go env GOPATH)" -name zarc -type f 2>/dev/null | head -1)
fi

if [ -n "$ZARC_BIN" ] && [ -f "$ZARC_BIN" ]; then
  echo ""
  echo " Executando zarc setup..."
  echo ""
  "$ZARC_BIN" setup --no-alias
else
  warn "Binário zarc não encontrado — rode 'zarc setup' manualmente após reiniciar o terminal"
fi

echo ""
echo " ─────────────────────────────────────────────"
echo " Instalação concluída!"
echo ""
echo " Reinicie o terminal e execute: zarc"
echo ""
