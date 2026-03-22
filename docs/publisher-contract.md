# Contrato do Publisher — Watchdog Monitor

## Transporte

| Propriedade | Valor |
|---|---|
| Protocolo | NATS (nats.go v1) |
| Encoding | Protocol Buffers (proto3) |
| QoS | At-most-once (NATS core, sem JetStream) |
| Intervalo de publicação | 5 segundos por host |

---

## Subject

```
watchdog.<hostname>.metrics
```

`<hostname>` é o nome da máquina sanitizado: apenas `[a-zA-Z0-9-]`, hífens consecutivos colapsados, sem hífen no início/fim.

**Exemplos:**
```
watchdog.DESKTOP-FH8A9CN.metrics
watchdog.srv-prod-01.metrics
```

**Wildcard para subscrever todos os hosts:**
```
watchdog.*.metrics
```

---

## Payload

Bytes binários serializados com `proto.Marshal` do seguinte schema:

```protobuf
syntax = "proto3";

message SystemMetrics {
  string hostname     = 1;   // nome sanitizado da máquina
  int64  timestamp_ms = 2;   // Unix timestamp em milliseconds UTC

  CpuMetrics    cpu    = 3;
  MemoryMetrics memory = 4;
  optional GpuMetrics gpu = 5;          // nil se GPU indisponível

  repeated NetworkInterface network       = 6;
  repeated ProcessInfo      top_processes = 7; // top 5 por CPU
}

message CpuMetrics {
  double          total_percent    = 1;  // 0.0 – 100.0
  repeated double per_core_percent = 2;  // um valor por core lógico
}

message MemoryMetrics {
  uint64 total_bytes     = 1;
  uint64 available_bytes = 2;
  double used_percent    = 3;  // 0.0 – 100.0
}

message GpuMetrics {
  double used_percent          = 1;  // sempre 0.0 (limitação WMI)
  uint64 dedicated_used_bytes  = 2;  // sempre 0 (limitação WMI)
  uint64 dedicated_total_bytes = 3;  // VRAM total em bytes (cap ~4 GB por uint32 WMI)
  optional double temperature_celsius = 4;  // nil se hardware não expõe
}

message NetworkInterface {
  string name             = 1;
  uint64 bytes_sent_delta = 2;  // bytes enviados desde o ciclo anterior
  uint64 bytes_recv_delta = 3;  // bytes recebidos desde o ciclo anterior
}

message ProcessInfo {
  string name         = 1;
  uint32 pid          = 2;
  double cpu_percent  = 3;
  uint64 memory_bytes = 4;  // RSS em bytes
}
```

O arquivo `.proto` e o código Go gerado estão em `proto/metrics.proto` e `gen/metrics/`.

---

## Comportamento garantido

| Situação | Comportamento |
|---|---|
| GPU indisponível | Campo `gpu` ausente no payload (nil) — nunca aborta o ciclo |
| Falha em qualquer coletor | Campo correspondente nil/zero — publish acontece normalmente |
| Falha no publish NATS | Erro logado e descartado — próximo ciclo tenta novamente |
| Primeiro ciclo de rede | `bytes_sent_delta` e `bytes_recv_delta` são `0` (sem baseline anterior) |
| Primeiro ciclo de CPU | `total_percent` pode ser `0.0` (sem baseline anterior) |

---

## Exemplo de subscriber em Go para Victoria Metrics

```go
nc, _ := nats.Connect("nats://192.168.21.X:4222")

nc.Subscribe("watchdog.*.metrics", func(msg *nats.Msg) {
    var payload metrics.SystemMetrics
    if err := proto.Unmarshal(msg.Data, &payload); err != nil {
        log.Printf("unmarshal error: %v", err)
        return
    }

    ts := time.UnixMilli(payload.TimestampMs)
    host := payload.Hostname

    // CPU
    fmt.Fprintf(w, "watchdog_cpu_total_percent{host=%q} %f %d\n",
        host, payload.Cpu.TotalPercent, ts.UnixMilli())

    // Memória
    fmt.Fprintf(w, "watchdog_memory_used_percent{host=%q} %f %d\n",
        host, payload.Memory.UsedPercent, ts.UnixMilli())
    fmt.Fprintf(w, "watchdog_memory_available_bytes{host=%q} %d %d\n",
        host, payload.Memory.AvailableBytes, ts.UnixMilli())

    // GPU (campo opcional — checar nil antes de acessar)
    if payload.Gpu != nil {
        fmt.Fprintf(w, "watchdog_gpu_dedicated_total_bytes{host=%q} %d %d\n",
            host, payload.Gpu.DedicatedTotalBytes, ts.UnixMilli())
        if payload.Gpu.TemperatureCelsius != nil {
            fmt.Fprintf(w, "watchdog_gpu_temperature_celsius{host=%q} %f %d\n",
                host, *payload.Gpu.TemperatureCelsius, ts.UnixMilli())
        }
    }

    // Rede
    for _, iface := range payload.Network {
        fmt.Fprintf(w, "watchdog_net_bytes_sent_delta{host=%q,iface=%q} %d %d\n",
            host, iface.Name, iface.BytesSentDelta, ts.UnixMilli())
        fmt.Fprintf(w, "watchdog_net_bytes_recv_delta{host=%q,iface=%q} %d %d\n",
            host, iface.Name, iface.BytesRecvDelta, ts.UnixMilli())
    }

    // Top processos
    for _, proc := range payload.TopProcesses {
        fmt.Fprintf(w, "watchdog_process_cpu_percent{host=%q,pid=\"%d\",name=%q} %f %d\n",
            host, proc.Pid, proc.Name, proc.CpuPercent, ts.UnixMilli())
        fmt.Fprintf(w, "watchdog_process_memory_bytes{host=%q,pid=\"%d\",name=%q} %d %d\n",
            host, proc.Pid, proc.Name, proc.MemoryBytes, ts.UnixMilli())
    }
})
```

---

## Pontos de atenção para o subscriber

- **`gpu` pode ser nil** — checar antes de acessar qualquer campo
- **`temperature_celsius` pode ser nil** dentro de `GpuMetrics` — desreferenciar com cuidado
- **`dedicated_used_bytes` é sempre 0** — WMI não expõe uso em tempo real; não criar métrica para esse campo
- **Primeiro ciclo de rede tem deltas 0** — considerar filtrar na ingestão inicial
- **Timestamps já chegam em UTC milliseconds** — converter com `time.UnixMilli()`
- **PIDs podem reutilizar valores** entre ciclos — não usar PID como identificador único de longo prazo
