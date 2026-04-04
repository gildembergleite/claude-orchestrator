
## Memória de Sessão

- **Ao iniciar qualquer nova sessão**, verificar se existe um arquivo de memória para o diretório/repositório atual no sistema de memória (`~/.claude/projects/<project-path>/memory/`).
- Se **não existir**, criar imediatamente um arquivo de memória do tipo `project` com o contexto inicial do repositório (nome do projeto, stack, objetivo geral, branch atual, etc.).
- **A cada iteração relevante na sessão**, atualizar o arquivo de memória do projeto com o que foi feito, decisões tomadas, e contexto importante — garantindo que sessões futuras nunca percam o histórico do que já foi realizado.
- O objetivo é manter continuidade entre sessões: qualquer nova conversa deve poder retomar de onde a anterior parou, sem que o usuário precise re-explicar o contexto.
