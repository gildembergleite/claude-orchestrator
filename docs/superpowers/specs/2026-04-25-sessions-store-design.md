# Sessions Store v1 — Design

**Date:** 2026-04-25
**Topic:** Schema rico e persistência segura para `~/.config/claude-orchestrator/sessions.json`
**Driver:** Brainstorming 2026-04-25 (item B). Várias features pendentes do `TODO.md` dependem de metadata por sessão (status, resumo, auto-cleanup, prompt inicial, workspaces). O schema atual `map[string]string` (nome → dir) bloqueia qualquer evolução.

## Goal

Substituir o schema raso atual por um store versionado, com metadata por sessão, escrita atômica, lock cooperativo entre instâncias e respeito a `XDG_CONFIG_HOME`. Manter migração silenciosa do schema antigo. Não introduzir features novas neste passo — só destravar a base.

## Current state and problems

`internal/tmux/sessions.go` hoje:

- Schema: `map[string]string` em JSON puro (sem `version`).
- Path: `~/.config/claude-orchestrator/sessions.json` hardcoded (não respeita `XDG_CONFIG_HOME`).
- Permissões: `0644` no arquivo, `0755` no dir — sessions.json pode conter paths sensíveis (`/private/`, `/etc/`, paths de credenciais).
- Escrita não atômica: `os.WriteFile` direto. Crash a meio caminho corrompe.
- Sem locking: duas instâncias do TUI rodando em paralelo (multi-tmux, multi-pane) sobrescrevem-se sem coordenação.
- `LoadSessions` retorna `(nil, nil)` quando arquivo não existe — confunde "vazio" com "erro silenciado".
- Tipo `tmux.Session` em `client.go` é `struct { Name string }` (resultado de `tmux ls`); colide com a ideia de uma "sessão registrada com metadata".

## New schema (v1)

```json
{
  "version": 1,
  "sessions": {
    "<name>": {
      "dir": "/abs/path",
      "created_at": "2026-04-25T10:00:00Z",
      "last_attached_at": "2026-04-25T15:30:00Z",
      "command": "",
      "env": {},
      "tags": [],
      "workspace": ""
    }
  }
}
```

**Required fields:** `dir`, `created_at`, `last_attached_at`.
**Optional (omitempty na serialização):** `command`, `env`, `tags`, `workspace`.

Princípio: **só persistir o que o usuário declarou**. Nada de status volátil (claude ativo/idle, branch git atual) — isso é runtime e fica em código separado.

### Por que map keyed by name e não array?
- Lookup O(1) por nome (operação dominante: "attach this session").
- Nome é a chave natural pro tmux (uma sessão por nome).
- Array exigiria índice paralelo ou loop linear em cada operação.

### Por que campos opcionais hoje?
- `command`/`env`: destrava futura "prompt inicial ao criar sessão" (TODO alta prioridade) sem nova migração.
- `tags`/`workspace`: destrava futura "workspaces YAML" e filtros no TUI.
- Custo de incluir agora: zero (omitempty + valor zero quando ausente).

## Public API (Go)

```go
package tmux

type Session struct {
    Name           string            `json:"-"` // populado da chave do map
    Dir            string            `json:"dir"`
    CreatedAt      time.Time         `json:"created_at"`
    LastAttachedAt time.Time         `json:"last_attached_at"`
    Command        string            `json:"command,omitempty"`
    Env            map[string]string `json:"env,omitempty"`
    Tags           []string          `json:"tags,omitempty"`
    Workspace      string            `json:"workspace,omitempty"`
}

// Package-level helpers (mantém compat com chamadores atuais)
func RegisterSession(name, dir string) error    // cria; created_at = last_attached_at = now()
func UnregisterSession(name string) error
func TouchSession(name string) error             // last_attached_at = now() (idempotente; no-op se não registrada)
func GetSession(name string) (Session, bool, error)
func ListRegistered() ([]Session, error)         // ordenada por last_attached_at desc
```

Nota: a API existente expõe `RegisterSession` e `UnregisterSession`. Mantemos a assinatura para minimizar blast radius nos callers (`cmd/claude-orchestrator/main.go`, `internal/tui/model.go`). Internamente, ambas chamam um `store` privado.

## Renomear o tipo `Session` do `client.go`

Hoje `client.ListSessions` retorna `[]Session{Name string}` (resultado de `tmux ls`). Pra dar o nome `Session` ao struct rico do store sem ambiguidade, mudamos `ListSessions` pra retornar `[]string` (lista de nomes). Os consumidores (`internal/tui/model.go`, `internal/tmux/client_test.go`) iteram por nome — perda zero de informação, ganho de clareza.

## Migration (legacy → v1)

Carga implementa detecção:

1. Ler arquivo. Se não existe, retornar store vazio v1.
2. Tentar `json.Unmarshal` no shape v1 (`{version, sessions}`). Se `version >= 1` e `sessions` é objeto, aceitar.
3. Caso contrário, tentar `json.Unmarshal` no shape legacy `map[string]string`. Se sucesso:
   - Para cada `(name, dir)`, criar `Session{Dir: dir, CreatedAt: now, LastAttachedAt: now}`.
   - Persistir imediatamente no shape v1 (regrava o arquivo).
4. Caso contrário, retornar erro com a mensagem original do unmarshal v1.

Resultado: usuários existentes não percebem a migração — abrem o TUI uma vez e o arquivo já vira v1.

## Concurrency & atomicity

### Atomic write
1. Escrever JSON em `sessions.json.tmp` no mesmo diretório (mesmo filesystem → rename atômico em POSIX).
2. `f.Sync()` antes de fechar.
3. `os.Rename(tmp, real)`.

Falha em qualquer etapa: arquivo `.tmp` é descartado, original intacto.

### Cooperative locking
Sidecar `~/.config/claude-orchestrator/sessions.lock` (arquivo vazio).
Operações que mutam (`Register`/`Unregister`/`Touch`):

```go
fd, _ := os.OpenFile(lockPath, os.O_CREATE|os.O_RDWR, 0600)
syscall.Flock(int(fd.Fd()), syscall.LOCK_EX) // bloqueia
defer syscall.Flock(int(fd.Fd()), syscall.LOCK_UN)
defer fd.Close()
// load → mutate → save
```

`syscall.Flock` é Unix-only. Goreleaser builda darwin/linux apenas — coberto. Não introduzimos suporte a Windows neste design.

`Get`/`List` são read-only; podem ler sem lock (worst case: snapshot levemente desatualizado, aceitável).

## Path resolution

```go
func configDir() string {
    if x := os.Getenv("XDG_CONFIG_HOME"); x != "" {
        return filepath.Join(x, "claude-orchestrator")
    }
    home, _ := os.UserHomeDir()
    return filepath.Join(home, ".config", "claude-orchestrator")
}
```

`MkdirAll(configDir(), 0700)` antes de escrever. Arquivo final em modo `0600`.

## Wiring

### `cmd/claude-orchestrator/main.go`

- `case "new"`: já chama `tmux.RegisterSession(name, dir)`. Sem mudança.
- `case "attach"`: adicionar `tmux.TouchSession(name)` antes do switch/attach. Falha de touch é silenciosa (não bloqueia o attach).

### `internal/tui/model.go`

- Trocar `sessions []tmux.Session` por `sessions []string` (tipo do client mudou).
- `UnregisterSession(name)` segue idêntico.

### `internal/tmux/client.go`

- Remover struct `Session`. `ListSessions` retorna `([]string, error)`.

## Out of scope (próximas etapas, registradas)

- **Auto-cleanup de sessões órfãs**: usar `LastAttachedAt` + verificação `os.Stat(Dir)` pra detectar diretório que sumiu. Feature separada, depende deste schema.
- **Status no menu** (claude ativo/idle, branch): runtime, separado do store.
- **Workspaces YAML**: leitor de `.claude-orchestrator.yml` que popula `Workspace` e `Tags`. Separado.
- **Prompt inicial**: usa `Command` no schema. Separado.

## Files touched

| Arquivo | Mudança |
|---|---|
| `internal/tmux/sessions.go` | Reescrita: novo schema, atomic write, flock, XDG, migração |
| `internal/tmux/sessions_test.go` | Reescrita: cobre roundtrip, migração legacy, XDG, modo 0600, atomic, locking |
| `internal/tmux/client.go` | Remove struct `Session`; `ListSessions` retorna `[]string` |
| `internal/tmux/client_test.go` | Ajusta expectativa para `[]string` |
| `internal/tui/model.go` | Tipo `sessions []string`; loop ajustado |
| `cmd/claude-orchestrator/main.go` | Adiciona `TouchSession` no `case "attach"` |

## Testing

- **Roundtrip**: serializar → desserializar → diff zero.
- **Migration**: arquivo `{"foo":"/bar"}` (legacy) carrega como Store v1 com timestamps populados; arquivo é regravado em v1.
- **Empty / missing file**: retorna store vazio v1 sem erro.
- **Permissions**: arquivo escrito com modo `0600`; dir com `0700`.
- **XDG_CONFIG_HOME**: variável seta caminho diferente; testes usam `t.Setenv`.
- **Atomic write**: tmp file não fica residual após sucesso.
- **Concurrency**: 50 goroutines fazendo `Register/Touch/Unregister` em paralelo; nenhum erro, JSON final válido.
- **Race detector**: rodar `go test -race ./internal/tmux/...`.

## Risks

- **Locking corrompe se filesystem não suporta `flock`** (raro: NFS sem `nfslock`). Mitigação: documentar; comportamento degradado é "race entre instâncias", não corrupção.
- **Goroutines em testes podem flakar em CI**. Mitigação: timeout generoso e checagem só do invariante final (arquivo válido + número correto de entradas).
- **Migração silenciosa**: sem aviso ao usuário. Aceitável: schema antigo é informacionalmente subset do novo, nenhum dado é perdido. Se virar incômodo no futuro, fácil adicionar log/flag.

## Acceptance criteria

- [ ] Build verde, `go test -race ./...` passa.
- [ ] Usuário com `sessions.json` antigo abre o TUI uma vez → arquivo vira v1, registros preservados, timestamps populados.
- [ ] Modo do arquivo após qualquer operação: `0600`.
- [ ] `TouchSession` chamada no attach atualiza `last_attached_at`.
- [ ] `XDG_CONFIG_HOME` honrado.
- [ ] Nenhum caller externo do pacote `tmux` quebra (assinatura de `RegisterSession`/`UnregisterSession` mantida).
