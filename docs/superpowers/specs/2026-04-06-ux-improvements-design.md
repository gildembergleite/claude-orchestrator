# UX Improvements: Alias, Dir Browser, Default Dir

**Data:** 2026-04-06
**Branch:** feat/bootstrap

## Resumo

Três melhorias de UX para o CLI zarc:

1. Alias personalizado durante o setup
2. Navegação de diretórios visual (file browser) estilo fish/zsh
3. Diretório atual como default para nova sessão

---

## 1. Alias personalizado no setup

### Fluxo

Durante o passo 5 do setup (shell config), antes de criar o alias, exibir um menu:

```
Como deseja chamar o CLI?
  > zarc + claude (dois aliases)
    Somente zarc
    Nome personalizado...
```

- **Opção 1:** Cria alias `zarc` e `claude`, ambos apontando para o binário
- **Opção 2:** Cria só `zarc` (comportamento atual)
- **Opção 3:** Abre input para o usuário digitar o nome, cria só esse alias

### Implementação

- `internal/setup/shell.go` — a função de configuração do shell recebe `[]string` de aliases ao invés de hardcoded `"zarc"`
- `internal/setup/setup.go` — coleta a escolha do usuário antes de chamar a configuração do shell
- **Fish:** cria um arquivo de função por alias em `~/.config/fish/functions/`
- **Bash/Zsh:** adiciona uma linha `alias` por nome no rc file

### Arquivos impactados

- `internal/setup/shell.go`
- `internal/setup/setup.go`

---

## 2. File browser integrado ao input de diretório

### Componente

Novo componente `DirBrowserModel` — híbrido com campo de texto no topo e lista de diretórios abaixo:

```
Diretório: ~/workspace/zarc-claude-orchestrator
─────────────────────────────────────────────────
  > projeto-a/
    projeto-b/
    projeto-c/
    outro-dir/
```

### Comportamento

- **Setas cima/baixo:** navega na lista de diretórios
- **Enter no item selecionado:** entra no subdiretório (drill-down), atualiza o campo de texto e lista os filhos
- **Backspace com campo vazio (ou Left no início):** volta ao diretório pai
- **Digitando:** filtra a lista por prefixo
- **Enter sem seleção na lista (ou Ctrl+D):** confirma o diretório atual exibido no campo
- **Esc:** cancela e volta ao menu principal
- Diretórios ocultos (`.`) ficam escondidos por padrão
- Máximo ~10 itens visíveis com scroll

### Implementação

- Novo arquivo `internal/tui/dirbrowser.go` com `DirBrowserModel`
- `internal/tui/model.go` usa `DirBrowserModel` no estado `stateNewSessionDir` ao invés de `InputModel`
- `InputModel` continua existindo para outros inputs (nome da sessão, alias customizado)

### Validação

Ao confirmar, verifica se o caminho existe e é um diretório (mesma lógica atual).

### Arquivos impactados

- `internal/tui/dirbrowser.go` (novo)
- `internal/tui/model.go`

---

## 3. Diretório atual como default

### Comportamento

Ao entrar no estado `stateNewSessionDir`, o file browser vem pré-preenchido com `os.Getwd()`:

```
Diretório: ~/workspace/zarc-claude-orchestrator
─────────────────────────────────────────────────
  > subdir-a/
    subdir-b/
```

- **Enter imediato:** aceita o diretório atual (`pwd`), vai direto para input de nome da sessão
- Nome da sessão continua pré-preenchido com `filepath.Base(dir)`
- Se o usuário navegar (setas, digitar), o default é substituído pela navegação

### Implementação

- `internal/tui/model.go` passa `os.Getwd()` como valor inicial do `DirBrowserModel`
- `DirBrowserModel` recebe `initialDir string` no construtor e lista o conteúdo desse diretório
- Campo de texto exibe caminho completo (convertido para `~/...` se dentro do home)

### Arquivos impactados

- `internal/tui/model.go` (inicialização do componente)
- `internal/tui/dirbrowser.go` (construtor com `initialDir`)

---

## Arquivos totais impactados

| Arquivo | Ação |
|---------|------|
| `internal/tui/dirbrowser.go` | Novo |
| `internal/tui/model.go` | Modificado |
| `internal/setup/shell.go` | Modificado |
| `internal/setup/setup.go` | Modificado |

## Fora de escopo

- Navegação fuzzy (só prefixo por agora)
- Mostrar diretórios ocultos
- Bookmarks de diretórios favoritos
