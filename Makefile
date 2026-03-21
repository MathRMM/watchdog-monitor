# =============================================================================
# Watchdog Monitor
# =============================================================================

.PHONY: help build proto test install-service

VERSION ?= dev

# Exibe os comandos disponíveis
help:
	@echo ""
	@echo "  Comandos disponíveis:"
	@echo ""
	@echo "  Build"
	@echo "    make build            Cross-compila watchdog.exe para Windows amd64"
	@echo "    make proto            Gera código Go a partir de proto/metrics.proto"
	@echo ""
	@echo "  Testes"
	@echo "    make test             Executa todos os testes"
	@echo ""
	@echo "  Deploy"
	@echo "    make install-service  Exibe instrução para registrar o serviço no SCM"
	@echo ""

# -----------------------------------------------------------------------------
# Build
# -----------------------------------------------------------------------------

# Cross-compila o binário para Windows amd64.
# A versão é injetada via ldflags em tempo de compilação.
build:
	GOOS=windows GOARCH=amd64 go build \
		-ldflags "-X main.Version=$(VERSION)" \
		-o bin/watchdog.exe \
		./cmd/watchdog/

# -----------------------------------------------------------------------------
# Proto (adicionado na Fase 2)
# -----------------------------------------------------------------------------
proto:
	@echo "[TODO] Target 'proto' será implementado na Fase 2"

# -----------------------------------------------------------------------------
# Testes
# -----------------------------------------------------------------------------
test:
	go test ./...

# -----------------------------------------------------------------------------
# Deploy / Install
# -----------------------------------------------------------------------------

# Exibe o comando sc create para registrar o serviço no Windows SCM.
# NÃO executa automaticamente — copie e execute no terminal Windows com privilégios de Administrador.
# Use uma conta de serviço dedicada com privilégios mínimos (SEC02).
#
# Exemplo de criação de conta de serviço:
#   net user WatchdogSvc <senha> /add
#   sc create WatchdogMonitor binPath= "C:\Servicos\watchdog\watchdog.exe" ^
#       obj= ".\WatchdogSvc" password= "<senha>" ^
#       start= auto DisplayName= "Watchdog Monitor"
#   sc description WatchdogMonitor "Coleta métricas do sistema e publica via NATS"
install-service:
	@echo ""
	@echo "  Para registrar o Watchdog Monitor como Windows Service, execute no"
	@echo "  terminal Windows (CMD ou PowerShell) com privilégios de Administrador:"
	@echo ""
	@echo "  sc create WatchdogMonitor binPath= \"C:\\Servicos\\watchdog\\watchdog.exe\" ^"
	@echo "      obj= \".\\WatchdogSvc\" password= \"<senha>\" ^"
	@echo "      start= auto DisplayName= \"Watchdog Monitor\""
	@echo ""
	@echo "  ATENÇÃO: Use uma conta de serviço dedicada com privilégios mínimos (SEC02)."
	@echo "  Restrinja as permissões do watchdog.toml para somente leitura pelo serviço (SEC01)."
	@echo ""
