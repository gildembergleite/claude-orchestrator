
## Regras de memória claude-orchestrator

### Atualização obrigatória
Ao final de CADA resposta que envolva alteração de código, execução de comando,
build, teste, deploy ou qualquer ação no projeto:
- Atualize o arquivo de memória do projeto com um resumo do que foi feito
- O resumo deve ser conciso: o que mudou, resultado (sucesso/falha), decisões tomadas
- Sobrescreva o resumo anterior — mantenha apenas o estado atual, não histórico

### Sessões claude-orchestrator
Para consultar outra sessão, leia ~/.config/claude-orchestrator/sessions.json para obter o diretório,
depois leia a memória e o git log recente desse diretório.
