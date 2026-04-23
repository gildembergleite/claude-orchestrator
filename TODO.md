# TODO

Roadmap de features para o claude-orchestrator.

## Alta prioridade — UX diária

### Quick switch com fuzzy finder
Comando direto para trocar de sessão sem abrir a TUI completa:
```bash
claude-orchestrator switch
```
Fuzzy search estilo `fzf` ou atalho tmux melhorado (`prefix + s`).

### Status das sessões no menu
Mostrar info contextual ao lado de cada sessão:
```
  backend    ~/workspace/api     main    ● claude rodando
  frontend   ~/workspace/app     feat/x  ○ idle 15min
```
Inclui: branch git, status do Claude (ativo/idle), tempo desde última atividade.

### Prompt inicial ao criar sessão
Permitir passar contexto direto para o Claude na criação:
```bash
claude-orchestrator new --dir ~/api --prompt "continue o módulo de notificações"
```
O Claude abre já com o contexto carregado.

### Modo efêmero (sem sessão persistente)
Permitir abrir o Claude sem criar uma sessão tmux persistente — útil para tarefas rápidas, perguntas pontuais ou exploração:
```bash
claude-orchestrator quick
# ou
claude-orchestrator --no-session
```
Abre o Claude diretamente no diretório atual sem registrar nada no `sessions.json` nem criar sessão tmux.

## Média prioridade — Produtividade

### Grupos de sessões (workspaces)
Arquivo `.claude-orchestrator.yml` na raiz do projeto define um conjunto de sessões:
```yaml
workspace: meu-saas
sessions:
  - name: backend
    dir: ./api
  - name: frontend
    dir: ./web
  - name: infra
    dir: ./terraform
```
Comandos:
- `claude-orchestrator up` — cria todas as sessões
- `claude-orchestrator down` — mata todas as sessões

### Resumo ao reconectar sessão
Ao fazer attach, mostrar o estado atual da sessão lendo a memória do projeto:
```
Última atividade: Criou módulo de notificações, 3 testes passando
Branch: feat/notifications (2 commits ahead)
```

### Auto-cleanup de sessões órfãs
Limpar automaticamente:
- Entradas no `sessions.json` cujo diretório não existe mais
- Sessões tmux mortas que ainda aparecem no registro
- Validação na inicialização do TUI

## Baixa prioridade — Nice to have

### Notificação cross-session
Quando um comando longo termina numa sessão, notificar as outras sessões ativas:
```
[backend] ✓ Build concluído — 45 testes passando
```

### Dashboard overview
Comando para visão geral de todas as sessões:
```bash
claude-orchestrator status
```
```
SESSÕES ATIVAS
  backend     main           ✓ 3 commits hoje
  frontend    feat/login     ✓ 1 commit hoje
  infra       main           ○ sem alterações
```

### Template de workspace compartilhável
Exportar/importar configuração de workspace para padronização entre equipe.
