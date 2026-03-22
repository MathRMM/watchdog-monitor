# NATS — Configuração no Homelab

**Serviço:** Watchdog Monitor
**Protocolo:** NATS (sem autenticação na v1 — campos preparados para futuro)
**Subject publicado:** `watchdog.<hostname-sanitizado>.metrics`
**Intervalo:** a cada 5 segundos
**Payload:** Protobuf binário (`SystemMetrics` — schema em `proto/metrics.proto`)

---

## 1. Instalar o NATS Server

### Docker (recomendado para homelab)

```bash
docker run -d \
  --name nats \
  --restart unless-stopped \
  -p 4222:4222 \
  -p 8222:8222 \
  nats:latest --http_port 8222
```

- **4222** — porta do cliente (usada pelo Watchdog)
- **8222** — HTTP de monitoramento (`http://<ip>:8222/varz`)

### Binário direto

```bash
# Instalar (Linux amd64)
curl -L https://github.com/nats-io/nats-server/releases/latest/download/nats-server-v2.x.x-linux-amd64.zip -o nats-server.zip
unzip nats-server.zip
sudo mv nats-server /usr/local/bin/

# Iniciar
nats-server --port 4222 --http_port 8222
```

---

## 2. Configurar o `watchdog.toml` no Windows

Edite o arquivo `watchdog.toml` na mesma pasta do `watchdog.exe`:

```toml
# URL do servidor NATS no homelab
nats_url = "nats://<IP-DO-HOMELAB>:4222"

# Caminho do arquivo de log (opcional — padrão: watchdog.log na pasta do exe)
log_path = "C:\\Servicos\\watchdog\\watchdog.log"

# Credenciais NATS — deixar vazio se sem autenticação (v1)
nats_user     = ""
nats_password = ""
```

Substitua `<IP-DO-HOMELAB>` pelo IP fixo ou hostname da máquina que roda o NATS.

---

## 3. Verificar conectividade

Do Windows (onde o Watchdog roda), testar antes de instalar o serviço:

```powershell
# Verificar porta alcançável
Test-NetConnection -ComputerName <IP-DO-HOMELAB> -Port 4222
```

---

## 4. Inspecionar mensagens recebidas

### Via NATS CLI (em qualquer máquina com acesso ao servidor)

```bash
# Instalar nats CLI
go install github.com/nats-io/natscli/nats@latest

# Assinar o subject do Watchdog (substitua <hostname> pelo hostname sanitizado do PC)
nats sub "watchdog.>" --server nats://<IP-DO-HOMELAB>:4222
```

O payload é Protobuf binário — para decodificar legível, use `protoc`:

```bash
nats sub "watchdog.>" --server nats://<IP-DO-HOMELAB>:4222 | \
  protoc --decode watchdog.metrics.SystemMetrics proto/metrics.proto
```

### Via monitoramento HTTP do servidor

```
http://<IP-DO-HOMELAB>:8222/connz   — conexões ativas
http://<IP-DO-HOMELAB>:8222/subsz   — subscriptions ativas
http://<IP-DO-HOMELAB>:8222/varz    — estatísticas gerais
```

---

## 5. Subjects publicados

| Subject | Descrição |
|---|---|
| `watchdog.<hostname>.metrics` | Payload `SystemMetrics` do host monitorado |

**Exemplo:** máquina com hostname `MY-PC` publica em `watchdog.MY-PC.metrics`.

Para subscrever todos os hosts de uma vez: `watchdog.>` ou `watchdog.*.*`.

---

## 6. Referência rápida

| Item | Valor |
|---|---|
| Porta padrão NATS | `4222` |
| Porta monitoramento HTTP | `8222` |
| Formato do payload | Protobuf v3 (`SystemMetrics`) |
| Schema | `proto/metrics.proto` |
| Intervalo de publicação | 5 segundos |
| Autenticação (v1) | nenhuma |
