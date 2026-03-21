# SPEC — Watchdog de Métricas do Sistema

**PRD de referência:** `.planning/prd_watchdog_metricas_sistema.md`
**Data:** 2026-03-21
**Status:** draft
**Autor:** Claude Code

---

## 1. Visão Geral Técnica

O Watchdog é um binário Go compilado para Windows (`GOOS=windows GOARCH=amd64`) que roda como Windows Service registrado no SCM. A arquitetura é pipeline simples e linear: ticker absoluto → coletores independentes → serialização Protobuf → publicação NATS.

**Decisões técnicas:**

- **Sem goroutines concorrentes por coletor:** coleta sequencial dentro do ciclo. Simples, previsível, sem race conditions. Se um coletor demorar, o próximo tick absorve (RN06 garante que não há drift acumulativo).
- **`gopsutil`** para CPU, RAM, rede e processos — biblioteca madura com suporte Windows nativo via PDH/WMI interno.
- **`go-ole` + `wmi`** para GPU AMD RX 6600 — queries WMI diretas para `Win32_VideoController` e eventuais namespaces AMD específicos.
- **`nats.go`** (cliente oficial NATS) com `Connect()` sem autenticação para v1. Reconexão automática habilitada — mas descarte silencioso em falha de `Publish()` conforme RN03.
- **`lumberjack`** para rotação de log — zero dependência de serviço externo.
- **`BurntSushi/toml`** para parsing de `watchdog.toml`.
- **`golang.org/x/sys/windows/svc`** para integração com SCM Windows.
- Payload serializado em Protobuf v3 (gerado via `protoc` + `protoc-gen-go`).

---

## 2. Tecnologias e Versões

| Tecnologia                          | Versão      | Uso na alteração                                        |
|-------------------------------------|-------------|---------------------------------------------------------|
| Go                                  | 1.22+       | Runtime e compilação do binário                         |
| `github.com/shirou/gopsutil/v3`     | v3.x        | Coleta de CPU, RAM, rede e processos                    |
| `github.com/go-ole/go-ole`          | v1.x        | Base para queries WMI (GPU)                             |
| `github.com/yusufpapurcu/wmi`       | v1.x        | Queries WMI tipadas para GPU AMD                        |
| `github.com/nats-io/nats.go`        | v1.x        | Cliente NATS para publicação                            |
| `google.golang.org/protobuf`        | v1.x        | Serialização Protobuf (runtime)                         |
| `github.com/BurntSushi/toml`        | v1.x        | Parsing do arquivo de configuração                      |
| `gopkg.in/natefinish/lumberjack.v2` | v2.x        | Rotação de arquivo de log                               |
| `golang.org/x/sys`                  | latest      | Integração com SCM Windows (svc package)                |
| Protocol Buffers (`protoc`)         | 3.x         | Geração de código Go a partir do `.proto`               |

---

## 3. Mapa de Arquivos

```
Watchdog Monitor/
├── [NEW] go.mod                                    — Módulo Go, dependências
├── [NEW] go.sum                                    — Lockfile de dependências
├── [NEW] watchdog.toml                             — Config de exemplo documentado
├── [NEW] Makefile                                  — Targets: build, proto, install-service
├── [NEW] proto/
│   └── [NEW] metrics.proto                         — Schema Protobuf do payload
├── [NEW] internal/
│   ├── [NEW] config/
│   │   ├── [NEW] config.go                         — Struct Config, função Load(path)
│   │   └── [NEW] config_test.go                    — Testes de parsing e validação
│   ├── [NEW] collector/
│   │   ├── [NEW] cpu.go                            — Coletor CPU geral + por core
│   │   ├── [NEW] cpu_test.go
│   │   ├── [NEW] memory.go                         — Coletor RAM (total, disponível, %)
│   │   ├── [NEW] memory_test.go
│   │   ├── [NEW] gpu.go                            — Coletor GPU via WMI (opcional)
│   │   ├── [NEW] gpu_test.go
│   │   ├── [NEW] network.go                        — Coletor delta bytes por interface
│   │   ├── [NEW] network_test.go
│   │   ├── [NEW] process.go                        — Top-5 processos por CPU
│   │   └── [NEW] process_test.go
│   ├── [NEW] publisher/
│   │   ├── [NEW] nats.go                           — Conexão NATS, Publish(), Close()
│   │   └── [NEW] nats_test.go
│   ├── [NEW] logger/
│   │   └── [NEW] logger.go                         — Setup lumberjack, funções de log
│   └── [NEW] service/
│       ├── [NEW] service.go                        — Handler SCM (Execute, Stop)
│       └── [NEW] cycle.go                          — Loop principal com ticker absoluto
├── [NEW] cmd/watchdog/
│   └── [NEW] main.go                               — Entry point, inicialização, svc.Run()
└── [NEW] gen/                                      — Código gerado pelo protoc (não editar)
    └── [NEW] metrics/
        └── [NEW] metrics.pb.go                     — Structs Go geradas do .proto
```

---

## 4. Fases de Implementação

---

### Fase 1 — Scaffold, Configuração e Windows Service

**Objetivo:** Ter um binário Go que inicia como Windows Service, lê `watchdog.toml` com sucesso (ou falha com erro claro), loga a versão e encerra corretamente quando o SCM envia stop.

**Arquivos desta fase:**
- `[NEW] go.mod` — módulo e dependências iniciais
- `[NEW] cmd/watchdog/main.go` — entry point com `svc.Run()`
- `[NEW] internal/config/config.go` — struct `Config`, função `Load(path string) (*Config, error)`
- `[NEW] internal/config/config_test.go` — testes de configuração
- `[NEW] internal/logger/logger.go` — setup de log rotacionado via lumberjack
- `[NEW] internal/service/service.go` — handler SCM: `Execute()`, canal de stop
- `[NEW] watchdog.toml` — arquivo de exemplo com comentários inline (RNF07)

**O que fazer:**

Criar a estrutura de módulo Go com todos os pacotes como diretórios vazios inicialmente. Implementar o parsing do `watchdog.toml` com a struct `Config` contendo ao menos o campo `nats_url`. Se o arquivo estiver ausente ou malformado, retornar erro descritivo e não subir o serviço (Alternativo D do PRD).

Implementar o handler de Windows Service usando `golang.org/x/sys/windows/svc`. O `Execute()` deve: aceitar o sinal de start, logar a versão do binário (RF14), e aguardar o sinal de stop do SCM para encerrar. O binário deve também aceitar execução direta (fora do SCM) para facilitar desenvolvimento.

Configurar o logger com lumberjack: rotação por tamanho (ex: 10 MB), retenção de 3 arquivos de backup. O caminho do arquivo de log pode ser relativo ao executável.

Atende: RF08, RF09, RF14, RN01 (estrutura base), RNF06, RNF07, SEC04 (campo credenciais preparado no toml).

**Testes (TDD):**

- [ ] `config` → dado arquivo `watchdog.toml` válido com `nats_url`, quando `Load()` é chamado, então retorna `*Config` com campo populado sem erro
- [ ] `config` → dado arquivo inexistente, quando `Load()` é chamado, então retorna erro contendo o caminho do arquivo
- [ ] `config` → dado arquivo com TOML malformado, quando `Load()` é chamado, então retorna erro descritivo (não panic)
- [ ] `config` → dado arquivo sem campo `nats_url`, quando `Load()` é chamado, então retorna erro de campo obrigatório ausente
- [ ] Caso de borda: arquivo `watchdog.toml` com campos extras desconhecidos deve ser aceito sem erro (forward compatibility)

**Checklist de validação:**

- [ ] `go build ./...` compila sem erros para `GOOS=windows GOARCH=amd64`
- [ ] `Load()` retorna erro claro quando arquivo ausente (RF09, Alternativo D)
- [ ] `Load()` retorna erro claro quando TOML inválido (RF09, Alternativo D)
- [ ] Logger grava em arquivo e rotaciona conforme configurado (RF13)
- [ ] `watchdog.toml` de exemplo contém comentários em todos os campos (RNF07)
- [ ] Todos os testes desta fase passando com `go test ./internal/config/...`
- [ ] Nenhum teste de fases anteriores quebrado

**Critério de avanço:** Esta fase está concluída quando todos os itens acima estiverem marcados. Somente então iniciar a Fase 2.

---

### Fase 2 — Schema Protobuf e Contratos de Payload

**Objetivo:** Ter o schema `.proto` definido, código Go gerado, e ser capaz de instanciar e serializar um `SystemMetrics` completo (com e sem campo GPU).

**Arquivos desta fase:**
- `[NEW] proto/metrics.proto` — definição do schema
- `[NEW] gen/metrics/metrics.pb.go` — código gerado (não editar manualmente)
- `[NEW] Makefile` — target `proto` que executa `protoc`

**O que fazer:**

Definir o arquivo `metrics.proto` seguindo o contrato da Seção 6 deste SPEC. Todos os campos de métricas que podem estar ausentes (GPU inteiro, temperatura da GPU) devem ser `optional` ou em mensagem separada para que a ausência seja distinguível do valor zero (RNF09).

Adicionar o target `proto` no Makefile que executa `protoc --go_out=gen --go_opt=paths=source_relative proto/metrics.proto`. O código gerado vai para `gen/metrics/`.

Não há lógica de negócio nesta fase — apenas definição de schema e verificação de que o código gerado serializa/deserializa corretamente.

Atende: RF06 (estrutura do payload), RF11, RF12, RNF09.

**Testes (TDD):**

- [ ] `payload` → dado um `SystemMetrics` com todos os campos preenchidos (incluindo GPU), quando serializado com `proto.Marshal()` e desserializado, então todos os valores são preservados
- [ ] `payload` → dado um `SystemMetrics` sem o campo `gpu` (nil), quando serializado, então a desserialização resulta em campo GPU nil (não zero-value)
- [ ] `payload` → dado um `GpuMetrics` sem `temperature_celsius`, quando serializado, então temperatura é distinguível de `0.0` após desserialização
- [ ] `payload` → dado `timestamp_ms` com valor Unix milliseconds UTC, quando serializado e desserializado, então valor é exato
- [ ] Caso de borda: payload com lista `top_processes` vazia serializa e desserializa sem erro

**Checklist de validação:**

- [ ] `make proto` gera `gen/metrics/metrics.pb.go` sem erros
- [ ] `go build ./...` compila incluindo o pacote gerado
- [ ] Campo `gpu` é `optional` (nil quando ausente, não zero-value) — RNF09
- [ ] Campo `temperature_celsius` dentro de `GpuMetrics` é `optional`
- [ ] Todos os testes desta fase passando com `go test ./...`
- [ ] Nenhum teste de fases anteriores quebrado

**Critério de avanço:** Esta fase está concluída quando todos os itens acima estiverem marcados. Somente então iniciar a Fase 3.

---

### Fase 3 — Coletores de Métricas

**Objetivo:** Ter 5 coletores independentes funcionando no Windows, cada um retornando os dados do PRD ou erro isolado — sem que a falha de um afete os demais.

**Arquivos desta fase:**
- `[NEW] internal/collector/cpu.go` + `cpu_test.go`
- `[NEW] internal/collector/memory.go` + `memory_test.go`
- `[NEW] internal/collector/gpu.go` + `gpu_test.go`
- `[NEW] internal/collector/network.go` + `network_test.go`
- `[NEW] internal/collector/process.go` + `process_test.go`

**O que fazer:**

Cada coletor expõe uma função `Collect() (*<Tipo>Metrics, error)`. A assinatura retorna erro em vez de panic — o ciclo principal decide o que fazer com o erro (RN02, RNF08).

**CPU (RF01):** Usar `gopsutil/cpu.Percent(0, false)` para total e `cpu.Percent(0, true)` para por-core. O intervalo `0` usa o delta interno do gopsutil.

**Memória (RF02):** Usar `gopsutil/mem.VirtualMemory()`. Extrair `Total`, `Available` e `UsedPercent`.

**GPU (RF03):** Query WMI em `Win32_VideoController` para `CurrentUsage` (se disponível) e `AdapterRAM`. Para temperatura, tentar namespace `root\wmi` com classe `MSAcpi_ThermalZoneTemperature` ou fallback AMD-específico. Se WMI retornar erro ou campos zerados, retornar `(nil, err)` — o ciclo omite o campo GPU do payload. Erros repetidos de WMI devem ser logados apenas uma vez (supressão de duplicatas — implementada na Fase 5, aqui apenas retornar o erro).

**Rede (RF04):** Usar `gopsutil/net.IOCounters(true)` para obter contadores por interface. O coletor deve manter estado interno (última leitura) para calcular o delta de bytes enviados/recebidos desde o ciclo anterior. Interfaces com status não-up são ignoradas. Interfaces com delta zero em ambas as direções são incluídas se estiverem ativas (caso de borda do PRD).

**Processos (RF05):** Usar `gopsutil/process.Processes()` para listar processos. Para cada um, obter `CPUPercent()`, `Name()` e `MemoryInfo()`. Ordenar por CPU decrescente e retornar os 5 primeiros. Se um processo encerrar durante a coleta, ignorar e prosseguir (caso de borda do PRD). O `CPUPercent()` do gopsutil requer uma chamada anterior para calcular o delta — documentar esse comportamento no código.

Atende: RF01, RF02, RF03, RF04, RF05, RN02, RNF08, RNF03 (coleta < 1s total).

**Testes (TDD):**

- [ ] `cpu` → dado sistema com ao menos 1 core, quando `Collect()` é chamado, então retorna `CpuMetrics` com `TotalPercent` entre 0 e 100 e `PerCorePercent` com len > 0
- [ ] `memory` → quando `Collect()` é chamado, então `TotalBytes > 0`, `AvailableBytes <= TotalBytes`, `UsedPercent` entre 0 e 100
- [ ] `gpu` → dado WMI indisponível ou retorno vazio, quando `Collect()` é chamado, então retorna `(nil, err)` sem panic
- [ ] `network` → dado dois `Collect()` consecutivos com tráfego entre eles, então `BytesSentDelta` e `BytesRecvDelta` refletem a diferença correta
- [ ] `network` → dado primeira chamada a `Collect()`, então deltas são 0 (sem estado anterior)
- [ ] `process` → quando `Collect()` é chamado, então retorna slice com no máximo 5 elementos, ordenados por CPU decrescente
- [ ] `process` → dado sistema com menos de 5 processos visíveis, então retorna todos sem erro
- [ ] Caso de borda `network`: interface que desaparece entre ciclos não causa panic
- [ ] Caso de borda `process`: processo que encerra durante coleta é ignorado, os demais são retornados normalmente

**Checklist de validação:**

- [ ] Cada coletor compila isoladamente para `GOOS=windows`
- [ ] `Collect()` nunca panics — todos os erros são retornados explicitamente (RNF08)
- [ ] Coletor de rede mantém estado entre chamadas para calcular delta (RF04)
- [ ] Coletor de GPU retorna `(nil, err)` quando WMI falha — não retorna zero-value (RF03, RN02)
- [ ] Top-5 processos ordenados por CPU decrescente (RF05, RN05)
- [ ] Todos os testes desta fase passando com `go test ./internal/collector/...`
- [ ] Nenhum teste de fases anteriores quebrado

**Critério de avanço:** Esta fase está concluída quando todos os itens acima estiverem marcados. Somente então iniciar a Fase 4.

---

### Fase 4 — Ciclo Principal e Publicação NATS

**Objetivo:** Ter o loop de 5 segundos sem drift conectando coletores → payload Protobuf → publicação NATS, com descarte silencioso em falha e subject correto com hostname sanitizado.

**Arquivos desta fase:**
- `[NEW] internal/publisher/nats.go` — `Publisher` struct com `Connect()`, `Publish()`, `Close()`
- `[NEW] internal/publisher/nats_test.go`
- `[NEW] internal/service/cycle.go` — `RunCycle()`: monta payload, publica, loga erros
- `[NEW] cmd/watchdog/main.go` — atualizado para instanciar publisher e iniciar ciclo

**O que fazer:**

**Publisher NATS:** Criar struct `Publisher` que encapsula a conexão NATS. `Connect(url string)` estabelece a conexão com reconexão automática habilitada (opção `nats.MaxReconnects(-1)`). `Publish(subject string, data []byte) error` publica o payload. Se `Publish()` falhar (NATS fora do ar), retorna erro — o ciclo descarta e loga (RN03). `Close()` fecha a conexão graciosamente.

**Sanitização de hostname (RN04, caso de borda do PRD):** No startup, obter hostname via `os.Hostname()`. Sanitizar: manter apenas `[a-zA-Z0-9-]`, substituir outros caracteres por hífen, remover hífens duplicados consecutivos. O subject NATS será `watchdog.{hostname_sanitizado}.metrics`.

**Ciclo principal com ticker absoluto (RF07, RN06):** Usar `time.NewTicker(5 * time.Second)`. O ticker do Go já agenda os ticks por tempo absoluto — ticks atrasados por coleta lenta não causam drift acumulativo. O `RunCycle()` é chamado a cada tick e executa: coletar todas as métricas → montar `SystemMetrics` com timestamp UTC em ms (RF12), hostname (RF11) → serializar com `proto.Marshal()` → publicar via `Publisher.Publish()`.

**Tratamento de falhas no ciclo:** Se coleta de GPU falhar → campo `gpu` fica nil no payload, log de erro emitido (supressão de duplicatas na Fase 5). Se publicação NATS falhar → log de erro e continua para o próximo tick (RN03). Se coleta de CPU/RAM/rede/processo falhar → campo correspondente fica com zero-value ou vazio, log de erro.

Atende: RF06, RF07, RF10, RF11, RF12, RN03, RN04, RN06.

**Testes (TDD):**

- [ ] `publisher` → dado NATS indisponível, quando `Publish()` é chamado, então retorna erro sem panic
- [ ] `publisher` → dado NATS disponível, quando `Publish()` é chamado com payload válido, então retorna nil e mensagem é entregue
- [ ] `hostname sanitization` → dado hostname `"MY-PC.local"`, então sanitizado é `"MY-PC-local"`
- [ ] `hostname sanitization` → dado hostname com espaços `"My PC"`, então sanitizado é `"My-PC"`
- [ ] `hostname sanitization` → dado hostname `"pc--lab"`, então hífens duplicados são colapsados para `"pc-lab"`
- [ ] `cycle` → dado todos os coletores retornando sucesso, quando `RunCycle()` é chamado, então `SystemMetrics` serializado contém todos os campos não-nulos
- [ ] `cycle` → dado coletor GPU retornando erro, quando `RunCycle()` é chamado, então payload é publicado com campo `gpu` nil
- [ ] `cycle` → dado `Publish()` retornando erro, quando `RunCycle()` retorna, então nenhum panic ocorre e ciclo prossegue
- [ ] Caso de borda: `timestamp_ms` no payload corresponde ao tempo UTC do momento da coleta (não do publish)

**Checklist de validação:**

- [ ] Ticker usa `time.NewTicker` — não `time.Sleep` (RN06, RF07)
- [ ] Subject NATS segue padrão `watchdog.{hostname}.metrics` com hostname sanitizado (RN04)
- [ ] Falha de `Publish()` não interrompe o serviço (RF10, RN03)
- [ ] Campo `gpu` é nil no payload quando coletor GPU falha (RN02, RNF09)
- [ ] `timestamp_ms` é Unix milliseconds UTC (RF12)
- [ ] `hostname` está presente em todos os payloads (RF11)
- [ ] Todos os testes desta fase passando com `go test ./internal/publisher/... ./internal/service/...`
- [ ] Nenhum teste de fases anteriores quebrado

**Critério de avanço:** Esta fase está concluída quando todos os itens acima estiverem marcados. Somente então iniciar a Fase 5.

---

### Fase 5 — Logging, Resiliência e Hardening

**Objetivo:** Fechar todos os requisitos de operação contínua: supressão de erros repetidos de GPU, versão no log de startup, tratamento gracioso de shutdown pelo SCM, e validação dos requisitos não-funcionais de resource usage.

**Arquivos desta fase:**
- `[MOD] internal/logger/logger.go` — adicionar supressão de erros duplicados
- `[MOD] internal/service/service.go` — shutdown gracioso: fechar publisher, parar ticker
- `[MOD] internal/service/cycle.go` — integrar supressão de erros de GPU (Alternativo C do PRD)
- `[MOD] cmd/watchdog/main.go` — logar versão do binário na inicialização (RF14)

**O que fazer:**

**Supressão de erros repetidos (Alternativo C do PRD):** Para o coletor GPU especificamente, implementar mecanismo de "log once" — o erro de WMI é logado na primeira ocorrência e suprimido nas subsequentes até que o coletor volte a funcionar (quando deve ser logado que a GPU voltou a responder). Usar um `sync.Once` ou flag de estado no ciclo. Isso evita poluição do log em máquinas sem GPU ou com WMI parcialmente indisponível.

**Versão do binário (RF14):** Usar `ldflags` do Go para injetar a versão no binário em tempo de compilação (`-ldflags "-X main.Version=1.0.0"`). Logar a versão no startup junto com a URL NATS configurada e o hostname sanitizado.

**Shutdown gracioso:** Quando o SCM enviar sinal de stop, o `Execute()` deve: parar o ticker, aguardar o `RunCycle()` atual terminar (se estiver em execução), fechar a conexão NATS, fechar o logger. Usar um `context.Context` cancelável passado ao ciclo para coordenar o shutdown.

**Makefile completo:** Adicionar targets `build` (cross-compile para Windows), `proto` (geração Protobuf), `test` (`go test ./...`), e `install-service` (executa `sc create` com os parâmetros corretos — apenas documenta o comando, não executa automaticamente).

Atende: RF13, RF14, RNF04 (sem leaks via shutdown gracioso), RNF08, SEC02 (documentação do install-service com conta de serviço correta).

**Testes (TDD):**

- [ ] `logger` → dado mesmo erro logado 3 vezes consecutivas, então apenas 1 entrada aparece no log (supressão de duplicatas)
- [ ] `logger` → dado erro suprimido, quando coletor volta a funcionar (sem erro), então próximo erro diferente é logado normalmente
- [ ] `service shutdown` → dado sinal de stop do SCM, quando `Execute()` retorna, então publisher está fechado e ticker parado sem goroutine leak
- [ ] `version` → dado binário compilado com `-ldflags "-X main.Version=1.0.0"`, quando serviço inicia, então log de startup contém `"version=1.0.0"`
- [ ] Caso de borda: shutdown durante `RunCycle()` em execução não causa deadlock

**Checklist de validação:**

- [ ] Erros repetidos de GPU logados apenas uma vez (RF13, Alternativo C)
- [ ] Versão aparece no log de inicialização (RF14)
- [ ] Shutdown gracioso sem goroutine leak verificável com `goleak` ou análise manual (RNF04)
- [ ] `make build` gera binário `watchdog.exe` compilado para `GOOS=windows GOARCH=amd64`
- [ ] `make test` executa todos os testes e passa
- [ ] Makefile tem target `install-service` documentado com instrução `sc create` e conta de serviço (SEC02)
- [ ] `watchdog.toml` de exemplo tem campo preparado para credenciais NATS futuras (SEC04)
- [ ] Todos os testes desta fase passando com `go test ./...`
- [ ] Nenhum teste de fases anteriores quebrado

**Critério de avanço:** Esta fase está concluída quando todos os itens acima estiverem marcados. O projeto está pronto para deploy.

---

## 5. Ordem de Execução e Dependências entre Fases

```
Fase 1 → Fase 2 → Fase 3 → Fase 4 → Fase 5
```

- **Fase 2** depende da Fase 1 (módulo Go e estrutura de pacotes existindo)
- **Fase 3** depende da Fase 2 (tipos Protobuf usados nos coletores para montar sub-structs)
- **Fase 4** depende da Fase 3 (usa todos os coletores) e Fase 2 (serialização Protobuf)
- **Fase 5** depende da Fase 4 (modifica o ciclo e o service handler)

Nenhuma fase pode ser executada em paralelo — dependências são estritas e lineares.

---

## 6. Contratos e Interfaces

### Schema Protobuf (`proto/metrics.proto`)

```proto
syntax = "proto3";
package watchdog.metrics;
option go_package = "github.com/mathrmm/watchdog-monitor/gen/metrics";

message SystemMetrics {
  string hostname       = 1;
  int64  timestamp_ms   = 2;  // Unix milliseconds UTC
  CpuMetrics     cpu    = 3;
  MemoryMetrics  memory = 4;
  optional GpuMetrics gpu = 5;  // nil quando indisponível
  repeated NetworkInterface network      = 6;
  repeated ProcessInfo       top_processes = 7;
}

message CpuMetrics {
  double          total_percent    = 1;  // 0.0–100.0
  repeated double per_core_percent = 2;
}

message MemoryMetrics {
  uint64 total_bytes     = 1;
  uint64 available_bytes = 2;
  double used_percent    = 3;  // 0.0–100.0
}

message GpuMetrics {
  double used_percent          = 1;
  uint64 dedicated_used_bytes  = 2;
  uint64 dedicated_total_bytes = 3;
  optional double temperature_celsius = 4;  // nil quando não suportado
}

message NetworkInterface {
  string name             = 1;
  uint64 bytes_sent_delta = 2;  // delta desde último ciclo
  uint64 bytes_recv_delta = 3;
}

message ProcessInfo {
  string name          = 1;
  uint32 pid           = 2;
  double cpu_percent   = 3;
  uint64 memory_bytes  = 4;
}
```

### Interface dos Coletores (Go)

```go
// Cada coletor segue este padrão funcional:
func Collect() (*gen_metrics.XxxMetrics, error)

// Coletor de rede é stateful — deve ser instanciado como struct:
type NetworkCollector struct { /* estado da última leitura */ }
func NewNetworkCollector() *NetworkCollector
func (c *NetworkCollector) Collect() ([]*gen_metrics.NetworkInterface, error)
```

### Struct de Configuração (Go)

```go
type Config struct {
    NatsURL     string `toml:"nats_url"`      // obrigatório
    LogPath     string `toml:"log_path"`      // opcional, default: "watchdog.log"
    // Preparado para futuro (SEC04):
    NatsUser    string `toml:"nats_user"`     // opcional
    NatsPassword string `toml:"nats_password"` // opcional
}
```

---

## 7. Riscos Técnicos

| Risco | Fase afetada | Mitigação |
|---|---|---|
| AMD RX 6600 pode não expor `CurrentUsage` via `Win32_VideoController` | Fase 3 | Testar query em runtime; se retornar 0 ou erro, marcar campo como ausente e logar. Investigar namespace `root\cimv2` e AMD-specific WMI classes. Documentar limitação no `watchdog.toml` |
| `gopsutil` CPU percent requer intervalo entre leituras para calcular delta | Fase 3 | Usar `cpu.Percent(interval, ...)` com intervalo não-zero na primeira chamada, ou aceitar que o primeiro ciclo retorna 0% para CPU e documentar esse comportamento |
| Conta de serviço com privilégios mínimos pode não ter acesso a WMI de processos | Fase 5 | Documentar no Makefile as permissões mínimas necessárias. Testar com conta `Local Service` antes de restringir mais |
| `protoc` e `protoc-gen-go` precisam estar instalados no ambiente de build | Fase 2 | Documentar pré-requisitos no README/Makefile. Checar se o `.pb.go` gerado pode ser commitado para evitar dependência de `protoc` em CI simples |
| Processos protegidos do Windows podem retornar `Access Denied` ao coletar CPU% | Fase 3 | Tratar `Access Denied` como erro ignorável por processo — apenas pular e continuar para o próximo |

---

## 8. Rastreabilidade PRD → SPEC

| Requisito PRD | Atendido na fase | Observação |
|---|---|---|
| RF01 — CPU geral + por core | Fase 3 | `internal/collector/cpu.go` |
| RF02 — RAM (total, disponível, %) | Fase 3 | `internal/collector/memory.go` |
| RF03 — GPU via WMI (%, VRAM, temp) | Fase 3 | `internal/collector/gpu.go` — retorna nil em falha |
| RF04 — Delta bytes por interface de rede | Fase 3 | `internal/collector/network.go` com estado |
| RF05 — Top 5 processos por CPU | Fase 3 | `internal/collector/process.go` |
| RF06 — Publicar Protobuf no NATS | Fase 4 | `internal/publisher/nats.go` + `cycle.go` |
| RF07 — Ticker absoluto sem drift | Fase 4 | `time.NewTicker` em `cycle.go` |
| RF08 — Windows Service (SCM) | Fase 1 | `internal/service/service.go` |
| RF09 — Ler `watchdog.toml` | Fase 1 | `internal/config/config.go` |
| RF10 — Descartar silenciosamente em falha NATS | Fase 4 | Tratamento de erro em `cycle.go` |
| RF11 — Hostname em cada payload | Fase 4 | Campo `hostname` em `SystemMetrics` |
| RF12 — Timestamp UTC (ms) em cada payload | Fase 4 | Campo `timestamp_ms` em `SystemMetrics` |
| RF13 — Log rotacionado sem impactar ciclo | Fase 1 (setup) + Fase 5 (supressão) | `internal/logger/logger.go` |
| RF14 — Versão no log de inicialização | Fase 5 | `ldflags` + log no startup |
| RNF01 — CPU < 1% em média | Fase 3 + 4 | Validar após integração completa; gopsutil é leve |
| RNF02 — RAM < 50 MB | Fase 4 + 5 | Sem buffers acumulativos; shutdown gracioso evita leaks |
| RNF03 — Ciclo < 1s | Fase 3 | Coleta sequencial; WMI é o candidato mais lento |
| RNF04 — Sem memory leaks em 24h | Fase 5 | Shutdown gracioso, sem goroutines pendentes |
| RNF05 — Windows 10 64-bit | Fase 1 | `GOOS=windows GOARCH=amd64` |
| RNF06 — Binário autocontido | Fase 1 | Go compila estaticamente por padrão |
| RNF07 — `watchdog.toml` documentado | Fase 1 | Comentários inline no arquivo de exemplo |
| RNF08 — Falha em um coletor não impede os demais | Fase 3 + 4 | Cada `Collect()` retorna erro isolado |
| RNF09 — Campos opcionais no Protobuf | Fase 2 | `optional GpuMetrics gpu` e `optional double temperature_celsius` |
| RN01 — Um payload por ciclo, sem agregação | Fase 4 | `cycle.go` publica exatamente uma vez por tick |
| RN02 — Falha parcial: omitir campo, logar, publicar | Fase 4 + 5 | GPU nil + log once implementados |
| RN03 — Falha NATS: descartar, sem retry | Fase 4 | Retorno de erro de `Publish()` apenas logado |
| RN04 — Subject com hostname sanitizado | Fase 4 | Sanitização no startup em `main.go` |
| RN05 — Top 5 por CPU atual, empate arbitrário | Fase 3 | Sort por CPU descrescente, top 5 |
| RN06 — Ticker absoluto sem drift acumulativo | Fase 4 | `time.NewTicker` (não `time.Sleep`) |
| SEC01 — Permissões restritas no `watchdog.toml` | Fase 5 | Documentado no Makefile/install-service |
| SEC02 — Serviço não roda como SYSTEM | Fase 5 | Instruções de `sc create` com conta específica |
| SEC03 — Sem dados sensíveis em excesso no log | Fases 3–5 | Logar apenas nome, PID e % — não argumentos de linha de comando |
| SEC04 — `watchdog.toml` suporta credenciais NATS | Fase 1 | Campos `nats_user` e `nats_password` na struct Config |

---

## Histórico de Revisões

| Versão | Data       | Alteração       |
|--------|------------|-----------------|
| 1.0    | 2026-03-21 | Criação inicial |
