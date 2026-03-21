# PRD — Watchdog de Métricas do Sistema

**Data:** 2026-03-21
**Status:** draft

---

## 1. Contexto e Problema

Não existe hoje nenhum mecanismo de coleta contínua de métricas do computador local. Sem isso, é impossível responder perguntas como: "o jogo travou porque a GPU estava saturada?", "a RAM acabou durante aquela reunião?", "qual aplicativo estava consumindo mais CPU naquele momento?".

O objetivo é preencher essa lacuna com um agente leve que rode silenciosamente em segundo plano no Windows, coletando o estado do sistema a cada 5 segundos e publicando os dados num servidor NATS em homelab para consumo futuro por um dashboard.

---

## 2. Objetivo

Ter um binário Go rodando como serviço Windows que:
- Coleta métricas de CPU, memória, GPU AMD, rede e processos a cada 5 segundos
- Publica cada leitura via NATS em formato Protobuf
- Opera sem impacto perceptível no uso diário do computador (jogos, trabalho)
- Suporta execução contínua indefinida sem degradação

---

## 3. Requisitos Funcionais

| ID   | Requisito                                                                                                                                 | Prioridade |
|------|-------------------------------------------------------------------------------------------------------------------------------------------|------------|
| RF01 | Coletar percentual de uso geral de CPU e uso por core lógico                                                                              | Alta       |
| RF02 | Coletar uso total, disponível e percentual de memória RAM                                                                                 | Alta       |
| RF03 | Coletar uso de GPU (percentual), memória dedicada usada/total e temperatura — via WMI para GPU AMD RX 6600                                | Alta       |
| RF04 | Coletar bytes enviados e bytes recebidos por interface de rede ativa desde o último ciclo                                                 | Alta       |
| RF05 | Coletar os 5 processos com maior uso de CPU no momento da leitura, incluindo nome, PID, % CPU e memória consumida                        | Alta       |
| RF06 | Publicar cada leitura como uma mensagem Protobuf no subject NATS `watchdog.{hostname}.metrics`                                           | Alta       |
| RF07 | Executar o ciclo de coleta a cada 5 segundos com precisão de relógio (não acumulativo — drift não deve se acumular ao longo do tempo)    | Alta       |
| RF08 | Rodar como Windows Service registrado no SCM (Service Control Manager), iniciando automaticamente no boot                                | Alta       |
| RF09 | Ler o endereço do servidor NATS a partir de um arquivo de configuração externo `watchdog.toml`                                           | Alta       |
| RF10 | Se o NATS estiver indisponível no momento da publicação, descartar a leitura silenciosamente e seguir para o próximo ciclo               | Alta       |
| RF11 | Incluir o hostname da máquina em cada payload publicado                                                                                   | Alta       |
| RF12 | Incluir timestamp UTC preciso (nanosegundos ou millisegundos) em cada payload                                                            | Alta       |
| RF13 | Logar erros de coleta e de publicação em arquivo de log rotacionado, sem impactar o ciclo principal                                      | Média      |
| RF14 | Expor versão do binário consultável via log na inicialização do serviço                                                                   | Baixa      |

---

## 4. Requisitos Não Funcionais

| ID    | Requisito                                                                                                   | Categoria        |
|-------|-------------------------------------------------------------------------------------------------------------|------------------|
| RNF01 | O watchdog não deve consumir mais que 1% de CPU em média durante operação normal                           | Performance      |
| RNF02 | O watchdog não deve consumir mais que 50 MB de memória RAM em steady state                                 | Performance      |
| RNF03 | A coleta completa de um ciclo deve concluir em menos de 1 segundo (não bloquear o próximo tick)           | Performance      |
| RNF04 | Sem memory leaks: uso de memória deve ser estável após 24h de execução contínua                            | Disponibilidade  |
| RNF05 | Compatível com Windows 10 64-bit ou superior                                                               | Compatibilidade  |
| RNF06 | O binário deve ser autocontido — sem dependência de runtime externo além do sistema operacional            | Manutenibilidade |
| RNF07 | O arquivo `watchdog.toml` deve ser documentado com comentários inline explicando cada campo               | Manutenibilidade |
| RNF08 | Falhas em uma fonte de métrica (ex: WMI GPU indisponível) não devem impedir a publicação das demais        | Disponibilidade  |
| RNF09 | O payload Protobuf deve ter campos opcionais para métricas que possam estar ausentes (ex: GPU)             | Manutenibilidade |

---

## 5. Regras de Negócio

- **RN01:** Cada publicação NATS representa exatamente um ciclo de coleta. Não há agregação nem suavização de valores antes da publicação.
- **RN02:** Se a coleta de uma métrica falhar (ex: WMI retorna erro), o campo correspondente no payload é omitido (valor padrão Protobuf), e um erro é logado. A publicação ocorre normalmente com os dados disponíveis.
- **RN03:** Se a publicação NATS falhar (conexão recusada, timeout), a leitura é descartada sem retry. O próximo ciclo tentará publicar normalmente.
- **RN04:** O subject NATS `watchdog.{hostname}.metrics` usa o hostname real do Windows no momento da inicialização do serviço — não é configurável.
- **RN05:** Os "5 processos mais pesados" são medidos por percentual de CPU no ciclo atual. Em caso de empate, a ordem é arbitrária.
- **RN06:** O intervalo de 5 segundos é baseado em ticker de relógio — o próximo tick é agendado a partir do tempo absoluto, não do fim da coleta anterior, evitando drift acumulativo.

---

## 6. Requisitos de Segurança

- **SEC01:** O arquivo `watchdog.toml` pode conter credenciais NATS no futuro — deve ter permissões de leitura restritas ao usuário do serviço Windows (não world-readable).
- **SEC02:** O serviço Windows deve rodar como conta de serviço com privilégios mínimos necessários para WMI e leitura de processos — não como SYSTEM ou Administrator.
- **SEC03:** Nenhuma métrica coletada (nomes de processos, uso de rede) deve ser logada em nível de detalhe que exponha informações sensíveis além do necessário para diagnóstico.
- **SEC04:** A conexão NATS v1 (sem autenticação) é aceita para a v1 desta entrega. Quando o servidor NATS for configurado com autenticação, o `watchdog.toml` deve suportar campos de credenciais sem que seja necessário recompilar o binário.

---

## 7. Casos de Uso

### Fluxo Principal

1. O Windows inicia e o SCM sobe o serviço `WatchdogMonitor` automaticamente
2. O serviço lê `watchdog.toml` e obtém o endereço NATS
3. O serviço estabelece conexão com o servidor NATS em `192.168.1.19:4222`
4. A cada 5 segundos, o serviço coleta todas as métricas disponíveis
5. Serializa o payload em Protobuf
6. Publica no subject `watchdog.{hostname}.metrics`
7. Repete indefinidamente até o serviço ser parado

### Fluxos Alternativos

- **Alternativo A — NATS indisponível na inicialização:** O serviço sobe normalmente, continua coletando a cada 5s, descarta cada publicação falhada com log de erro, e tenta publicar novamente no próximo ciclo quando o NATS voltar.
- **Alternativo B — NATS cai durante operação:** Idêntico ao Alternativo A. Sem buffer, sem retry por ciclo.
- **Alternativo C — Falha de leitura WMI (GPU):** O campo GPU é omitido do payload. O restante das métricas é publicado normalmente. Erro é logado uma vez (não a cada ciclo para não poluir o log).
- **Alternativo D — `watchdog.toml` ausente ou malformado:** O serviço falha na inicialização com log de erro claro indicando o problema e o caminho esperado do arquivo. Não sobe.

### Casos de Borda

- **Múltiplas interfaces de rede:** Listar delta de bytes enviados/recebidos individualmente por interface ativa (ex: `Ethernet`, `Wi-Fi`). Interfaces com zero tráfego no ciclo são incluídas se estiverem ativas.
- **Máquina sem GPU ou WMI indisponível:** Omitir campos GPU do payload (RN02)
- **Hostname com caracteres especiais:** O subject NATS deve usar o hostname sanitizado (apenas alfanuméricos e hífens) para evitar subjects inválidos
- **Processo encerrado durante coleta:** Ignorar e prosseguir para o próximo processo na lista
- **Tick atrasado por coleta lenta:** O próximo tick é reagendado pelo tempo absoluto, não pelo fim da coleta — drift não se acumula (RN06)

---

## 8. Fora do Escopo

- Dashboard de visualização ou análise histórica
- Persistência local de dados (banco, arquivo)
- Alertas ou notificações baseados em thresholds
- Coleta de métricas de disco (I/O, uso por partição) — pode ser adicionado na v2
- Suporte a Linux ou macOS no cliente watchdog
- Configuração dinâmica sem reiniciar o serviço
- Criptografia TLS na conexão NATS (v1 — pode ser adicionado quando o servidor for configurado)
- Autenticação NATS na v1 (servidor ainda sem configuração de autenticação)
- Instalação automatizada do serviço Windows (o registro no SCM pode ser feito manualmente via `sc create` ou por script avulso)

---

## 9. Dependências e Riscos

| Tipo        | Descrição                                                                                                    | Impacto |
|-------------|--------------------------------------------------------------------------------------------------------------|---------|
| Dependência | Servidor NATS em `192.168.1.19:4222` precisa estar acessível na rede local                                 | Alto    |
| Dependência | WMI disponível e funcional no Windows alvo para coleta de GPU e métricas de sistema                        | Alto    |
| Risco       | AMD RX 6600 pode não expor temperatura via WMI padrão — pode exigir queries WMI específicas da AMD ou fallback sem temperatura | Médio   |
| Risco       | Leitura de % CPU por processo via WMI tem custo não negligenciável — pode impactar RNF01 se não otimizado  | Médio   |
| Risco       | Protobuf exige manutenção de arquivo `.proto` e geração de código — adiciona etapa no build                | Baixo   |
| Dependência | Configuração de autenticação NATS futura pode exigir atualização do `watchdog.toml` e do schema Protobuf  | Baixo   |
| Risco       | Conta de serviço com privilégios mínimos pode não ter acesso a todos os contadores WMI necessários         | Médio   |

---

## 10. Critérios de Aceitação

- [ ] O serviço inicia automaticamente com o Windows sem intervenção manual
- [ ] Mensagens Protobuf são publicadas no NATS a cada 5 segundos (±100ms de tolerância)
- [ ] O payload contém: timestamp UTC, hostname, métricas de CPU (geral + por core), memória, GPU (quando disponível), rede (por interface), top 5 processos por CPU
- [ ] Com o NATS desligado, o serviço permanece rodando sem crash e retoma publicação quando o NATS volta
- [ ] Com WMI de GPU falhando, o serviço publica as demais métricas normalmente
- [ ] Após 1 hora de execução, o uso de memória do processo é estável (sem crescimento contínuo)
- [ ] O uso de CPU pelo watchdog não ultrapassa 1% em média durante operação normal
- [ ] O arquivo `watchdog.toml` ausente impede a inicialização com mensagem de erro legível no log
- [ ] O subject NATS segue o padrão `watchdog.{hostname}.metrics` com hostname sanitizado
- [ ] SEC02: o serviço não roda como SYSTEM ou Administrator

---

## Histórico de Revisões

| Versão | Data       | Alteração       |
|--------|------------|-----------------|
| 1.0    | 2026-03-21 | Criação inicial |
