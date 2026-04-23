# Session Aliases — Cross-Session Communication

**Data:** 2026-04-06
**Branch:** dev

## Resumo

Registro de sessões tmux em `~/.config/zarc/sessions.json` que permite ao Claude Code de uma sessão consultar o que foi feito em outra sessão. Combinado com instrução de memória automática no CLAUDE.md para manter o contexto sempre atualizado.

## Registro de sessões

### Arquivo: `~/.config/zarc/sessions.json`

```json
{
  "backend": "/Users/dev/workspace/api",
  "frontend": "/Users/dev/workspace/app"
}
```

### Ciclo de vida

- **Criar sessão tmux** → registra `nome: diretório` no JSON
- **Matar sessão tmux** → remove a entrada do JSON
- Idempotente: se o nome já existe, sobrescreve o diretório

### Consulta cross-session

O dev na sessão `frontend` diz: "veja o que foi feito na sessão backend"

O Claude:
1. Lê `~/.config/zarc/sessions.json` → encontra `backend: ~/workspace/api`
2. Lê `~/.claude/projects/<path>/memory/MEMORY.md` do diretório backend
3. Lê o `git log` recente do diretório backend

## Memória automática via CLAUDE.md

### Problema

O Claude não atualiza a memória consistentemente — depende do dev pedir.

### Solução

O `zarc setup` adiciona ao `~/.claude/CLAUDE.md` uma instrução comportamental que obriga o Claude a atualizar a memória a cada iteração.

### Instrução adicionada

```markdown
## Regras de memória zarc

### Atualização obrigatória
Ao final de CADA resposta que envolva alteração de código, execução de comando,
build, teste, deploy ou qualquer ação no projeto:
- Atualize o arquivo de memória do projeto com um resumo do que foi feito
- O resumo deve ser conciso: o que mudou, resultado (sucesso/falha), decisões tomadas
- Sobrescreva o resumo anterior — mantenha apenas o estado atual, não histórico

### Sessões zarc
Para consultar outra sessão, leia ~/.config/zarc/sessions.json para obter o diretório,
depois leia a memória e o git log recente desse diretório.
```

## Arquivos impactados

| Arquivo | Ação |
|---------|------|
| `internal/tmux/sessions.go` | Novo — CRUD do sessions.json (Register, Unregister, Load) |
| `internal/tmux/client.go` | Modificar — chamar Register ao criar sessão |
| `internal/tui/model.go` | Modificar — chamar Unregister ao matar sessão |
| `internal/setup/claude.go` | Modificar — adicionar instrução de memória/sessões ao CLAUDE.md |
| `configs/claude-sessions.md` | Novo — template embarcado da instrução |
| `configs/embed.go` | Modificar — adicionar embed do novo template |

## Fora de escopo

- Sincronização automática entre sessões (o dev pede manualmente)
- UI para listar sessões registradas (o JSON é suficiente)
- Histórico de iterações (apenas resumo atual)
