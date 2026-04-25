# Rich Menu + Orphan Auto-Cleanup — Design

**Date:** 2026-04-25
**Topic:** Enriquecer o menu principal do TUI com metadata por sessão e remover entradas registradas cujo diretório sumiu.
**Driver:** Brainstorming 2026-04-25 (item C). O store v1 já entrega `LastAttachedAt` e `Dir` por sessão; o menu hoje só mostra o nome. Aproveitamos a metadata pra dar contexto imediato e fazer housekeeping silencioso.

## Goal

1. Cada item do menu principal mostra: **nome**, **diretório colapsado em `~`**, **idle time** (tempo desde último attach).
2. Antes de renderizar, sessões registradas cujo `Dir` não existe mais são desregistradas silenciosamente. Não toca em sessões tmux vivas.

Não-goals (próximas features, registradas):
- Status do Claude (rodando/idle): exige inspeção de panes via `tmux list-panes`. Próximo design.
- Branch git por sessão: idem, exige `git rev-parse` por sessão. Próximo design.
- Mostrar registradas-mas-sem-tmux-vivo no menu: tmux-resurrect já cobre o reboot loop. Mantemos o menu refletindo só o que tmux conhece.

## Current state

`internal/tui/model.go` `loadMainMenu`:

```go
sessions, _ := tmux.ListSessions(m.tmuxBin)  // []string
items := []MenuItem{{Label: "[+] Nova sessão", ID: "new"}}
for _, name := range sessions {
    items = append(items, MenuItem{Label: name, ID: "session:" + name})
}
```

`MenuItem.Label` é renderizada como string única. Sem coluna, sem metadata.

`internal/tmux/sessions.go` v1 expõe `GetSession(name)` (com `Dir`/`LastAttachedAt`) e `ListRegistered()`. Não há função pra cleanup.

## Auto-cleanup

Adicionar a `internal/tmux/sessions.go`:

```go
// CleanupOrphans desregistra sessões cujo Dir não existe.
// Retorna a quantidade removida e erro de I/O se houver.
func CleanupOrphans() (int, error)
```

Lógica:
1. `load()` o store sob lock.
2. Para cada `(name, sess)` em `s.Sessions`: se `os.Stat(sess.Dir)` errar com `ErrNotExist`, remover.
3. Salvar.
4. Retornar contagem removida.

Chamada:
- `loadMainMenu` invoca `tmux.CleanupOrphans()` antes de listar — erro logado em `m.err` mas não interrompe.

## Menu enrichment

Em `loadMainMenu`, para cada live tmux session:

1. Tentar `tmux.GetSession(name)`.
2. Se encontrada, montar label com 3 colunas. Se não, label só com o nome.

### Label rendering

Formato fixo, alinhado por padding com espaços:

```
<name padded>  <dir padded>  <idle right-aligned>
```

- Nome: pad até `nameWidth` (largura do mais longo, mínimo 8).
- Dir: colapsado (`~/...`) e truncado à esquerda (`...api`) se passar de `dirWidth` (limite 36).
- Idle: alinhado à direita, largura fixa 12 (`idle 5m   `, `agora     `).

Sessão sem registro (não está em `sessions.json`): só nome, sem metadata.

### Helpers

Dois helpers privados em `internal/tui/format.go`:

```go
// collapsePath substitui prefix HOME por ~.
func collapsePath(p string) string

// formatIdle traduz duração em string compacta:
//   < 1min       → "agora"
//   < 1h         → "Xm"
//   < 24h        → "Xh"
//   < 7d         → "Xd"
//   ≥ 7d         → "Xw"
func formatIdle(d time.Duration) string

// truncateLeft mantém os últimos n chars com prefixo "...".
func truncateLeft(s string, max int) string
```

Cobertos por testes unitários puros (sem dependência de filesystem).

## Files touched

| Arquivo | Mudança |
|---|---|
| `internal/tmux/sessions.go` | + `CleanupOrphans()` |
| `internal/tmux/sessions_test.go` | + cobertura cleanup |
| `internal/tui/format.go` | novo: `collapsePath`, `formatIdle`, `truncateLeft` |
| `internal/tui/format_test.go` | novo |
| `internal/tui/model.go` | `loadMainMenu` chama cleanup, monta label rico |

## Testing

- `CleanupOrphans`: dir presente preserva entrada; dir ausente remove; outros erros de stat (permissão) **preservam** (não removem) — só `ErrNotExist` dispara remoção.
- `collapsePath`: HOME real, HOME inexistente, paths fora de HOME, path vazio.
- `formatIdle`: limites de cada bucket, durações negativas (futuro próximo) viram "agora".
- `truncateLeft`: preserva quando cabe, trunca com `...` quando excede.
- Testes existentes (46) seguem verdes.

## Risks

- **Cleanup mata entrada mid-edit**: usuário pode estar editando o `Dir` no shell e o snapshot do filesystem mostra ausência momentânea. Mitigação: só remover se `os.Stat` retornar `ErrNotExist` explícito; outros erros (EACCES, EBUSY) preservam.
- **Label cresce muito em telas estreitas**: largura total fixa ~60 chars cabe em 80-col terminal. Se quebrar, o `MenuModel` já trunca pela viewport do bubbletea — degradação grácil.
- **Idle "agora" em sessão antiga regista**: pode confundir. Aceito porque `RegisterSession` em overwrite atualiza `LastAttachedAt` (comportamento já especificado no store v1).

## Acceptance criteria

- [ ] Build verde, `go test -race ./...` passa.
- [ ] Sessão registrada com `Dir` ausente desaparece silenciosamente do menu na próxima abertura.
- [ ] Cada item de sessão registrada exibe `~/dir   idle Xm` (ou equivalente).
- [ ] Sessão tmux viva sem registro aparece com só o nome (sem quebra).
- [ ] Helpers `collapsePath`/`formatIdle`/`truncateLeft` testados isoladamente.
