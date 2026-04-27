# Scriptable CLI + Prompt Inicial + AltScreen — Design

**Date:** 2026-04-27
**Topic:** Adicionar subcomandos não-interativos (`list`/`new`/`attach`/`kill`), flag `--prompt` para passar contexto inicial ao Claude na criação de sessão, e habilitar tela alternativa no TUI interativo.
**Driver:** Brainstorming 2026-04-25 (item D). Hoje o tool é TUI-only — usuário não consegue automatizar nem integrar com outros scripts. `--prompt` destrava o "prompt inicial" do TODO (alta prio). AltScreen resolve gargalo de UX (TUI rabisca o histórico do terminal).

## Goal

1. **Subcomandos scriptáveis**:
   - `claude-orchestrator list [--json]` — lista sessões registradas
   - `claude-orchestrator new --dir <path> [--name <n>] [--prompt <p>]` — cria sessão sem TUI
   - `claude-orchestrator attach <name>` — attach direto
   - `claude-orchestrator kill <name>` — kill direto
2. **`--prompt`**: contexto passado ao Claude no startup. Vai como argumento posicional pra binário `claude`, com shell quoting seguro.
3. **AltScreen**: TUI interativo passa a usar `tea.WithAltScreen()` — sai limpo, não rabisca o terminal.

Não-goals (próximas etapas):
- `--workspace`/`--tag` em `new` (podem usar `Workspace`/`Tags` do store v1, mas postergamos pra design de "workspaces YAML").
- Watcher de arquivo `.claude-orchestrator.yml`.
- Status do Claude rodando/idle no `list`.

## Current state

`cmd/claude-orchestrator/main.go`:
- Root command roda TUI; única subcomando é `setup`.
- Não há saída structured (JSON), não há comandos não-interativos.
- TUI roda sem `tea.WithAltScreen()` (rabisca o histórico).

`internal/tmux/sessions.go`:
- `RegisterSession(name, dir)` aceita só dois args; não há jeito de setar `Command`/`Tags`/`Workspace` ao criar.

## API change: variadic options em RegisterSession

Mantém compat com chamadores atuais. Padrão functional options:

```go
type RegisterOption func(*Session)

func WithCommand(cmd string) RegisterOption       { ... }
func WithTags(tags ...string) RegisterOption      { ... }
func WithWorkspace(ws string) RegisterOption      { ... }
func WithEnv(env map[string]string) RegisterOption { ... }

// Compat: RegisterSession(name, dir) continua funcionando.
func RegisterSession(name, dir string, opts ...RegisterOption) error
```

Aplicação: opts são aplicados ao `Session` resultante após o setup base (Dir/CreatedAt/LastAttachedAt). Em **overwrite** (sessão já existe), o registro existente é a base; cada `RegisterOption` passada modifica o campo correspondente. Não passar uma option = preservar o valor atual da sessão. Passar `WithCommand("")` = limpar.

## Subcommands

### `list`

```
claude-orchestrator list           # tabela alinhada
claude-orchestrator list --json    # array JSON, um objeto por sessão
```

Pretty output (3 colunas): `<name>  <dir-colapsado>  <idle>`. Zero sessões registradas: imprime "(nenhuma sessão registrada)" em stderr e retorna exit 0.

JSON output: array de objetos com todos os campos do `Session` (CreatedAt/LastAttachedAt em RFC3339).

### `new`

```
claude-orchestrator new --dir <path> [--name <n>] [--prompt <p>]
```

- `--dir` obrigatório. Resolvido via `filepath.Abs`.
- `--name` default = `filepath.Base(dir)`.
- `--prompt`: se não vazio, embute no comando claude com shell quoting seguro.
- Falha se sessão tmux com mesmo nome já existe (mensagem clara, exit 1).
- Após criar: registra no store (`RegisterSession` com `WithCommand(prompt)` se setado), attach se não dentro de tmux / switch se dentro.

### `attach`

```
claude-orchestrator attach <name>
```

- Falha se sessão tmux com nome não existe.
- Chama `TouchSession(name)` antes do attach (mesma semântica do TUI).
- Comportamento attach vs switch baseado em `IsInsideTmux()`.

### `kill`

```
claude-orchestrator kill <name>
```

- `KillSession` + `UnregisterSession`. Idempotente: nome inexistente retorna OK.

## Shell quoting

Pra embutir `--prompt` no comando passado ao tmux (`bin new-session -d <command>`):

```go
// shellQuote envolve em aspas simples, escapando apóstrofos internos.
func shellQuote(s string) string {
    return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
```

Chamada: `claudeCmd := fmt.Sprintf("%s %s", claudeBin, shellQuote(prompt))`.

Cobre todos os caracteres especiais (espaços, `$`, `"`, `;`, etc.). Único caso patológico: prompt com `\0`, que tmux/shell já recusam.

## Testability

Bodies dos subcomandos extraídos como funções puras:

```go
func runList(out io.Writer, jsonMode bool) error
func runNew(args NewArgs) error           // recebe tmuxBin/claudeBin via args
func runAttach(tmuxBin, name string) error
func runKill(tmuxBin, name string) error
```

Cobra commands são wrappers thin que coletam flags e chamam essas funções. Testes unitários cobrem:
- `runList` com store vazio / com várias sessões / `--json`
- `shellQuote` com casos patológicos
- `RegisterSession` com `WithCommand` set
- `RegisterSession` overwrite: opts não passadas preservam o valor anterior; opts passadas sempre sobrescrevem (inclusive com valor zero)

## AltScreen

`cmd/claude-orchestrator/main.go::runTUI`:

```go
- p := tea.NewProgram(app)
+ p := tea.NewProgram(app, tea.WithAltScreen())
```

Comportamento esperado:
- TUI ocupa tela inteira durante interação
- Ao sair (qualquer caminho — quit, attach, kill, esc), tela alternativa some e o histórico anterior aparece intacto
- Quando o action é `attach`, o tmux take-over substitui o terminal (alt screen é descartada antes)

Sem flag pra desligar; comportamento padrão de TUIs modernos.

## Files touched

| Arquivo | Mudança |
|---|---|
| `internal/tmux/sessions.go` | + `RegisterOption`, `WithCommand`, `WithTags`, `WithWorkspace`, `WithEnv`; assinatura variádica |
| `internal/tmux/sessions_test.go` | + cobertura options |
| `cmd/claude-orchestrator/main.go` | + subcomandos `list`, `new`, `attach`, `kill`; tea.WithAltScreen; runTUI passa prompt para registerSession |
| `cmd/claude-orchestrator/runner.go` (novo) | corpo testável: `runList`, `runNew`, `runAttach`, `runKill`, `shellQuote` |
| `cmd/claude-orchestrator/runner_test.go` (novo) | tests unitários |

## Testing

- `RegisterSession` com `WithCommand("p")`: persiste no store, recuperável via `GetSession`.
- `RegisterSession` overwrite com nova `WithCommand`: substitui valor antigo.
- `shellQuote`: `hello`, `it's`, `$(rm -rf /)`, vazio, `\n`, multi-linha.
- `runList` JSON: marshal válido, todos os campos.
- `runList` pretty: linhas alinhadas, sessão sem registro mostra "(nenhuma...)" em stderr.
- 54 testes anteriores seguem verdes.

## Risks

- **Subcomandos colidem com sessions chamadas como flags**: cobra resolve nomes de comando antes de flags posicionais; conflito é improvável. Mitigação: documentar que nome de sessão começando com `-` deve usar `--`.
- **AltScreen + processo claude que escreve direto na tty antes do tmux take-over**: na prática, `tmux new-session -d` cria a sessão detached e `attach-session` faz o take-over depois. Sem race.
- **Sessão tmux criada manualmente fora do tool**: `attach`/`kill` funcionam (tmux só checa nome). `list` mostra apenas registradas — sessões "live, não registradas" não aparecem no `list` (consistente: `list` é do store, não de `tmux ls`).

## Acceptance criteria

- [ ] Build verde, `go test -race ./...` passa.
- [ ] `claude-orchestrator list` em store vazio retorna OK e nada em stdout.
- [ ] `claude-orchestrator list --json` em store com 2 sessões retorna array JSON válido.
- [ ] `claude-orchestrator new --dir /tmp/test --name foo --prompt "olá"` cria sessão tmux, registra com `Command="olá"`, attacha (ou switch se dentro de tmux).
- [ ] `claude-orchestrator attach foo` toca `LastAttachedAt`, attacha.
- [ ] `claude-orchestrator kill foo` mata sessão e desregistra.
- [ ] TUI interativo entra em alt screen e sai limpo.
- [ ] Subcomando `setup` original segue funcional.
