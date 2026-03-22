# =============================================================================
# Watchdog Monitor
# =============================================================================

.PHONY: help build proto test run logs install-service

VERSION ?= dev

# Exibe os comandos disponíveis
help:
	@echo ""
	@echo "  Comandos disponíveis:"
	@echo ""
	@echo "  Build"
	@echo "    make build              Cross-compila watchdog.exe para Windows amd64"
	@echo "    make proto              Gera código Go a partir de proto/metrics.proto"
	@echo ""
	@echo "  Testes"
	@echo "    make test               Executa todos os testes"
	@echo ""
	@echo "  Execução local (WSL2 / Windows Terminal)"
	@echo "    make run                Compila e executa watchdog.exe localmente"
	@echo "    make logs               Monitora o arquivo de log em tempo real"
	@echo ""
	@echo "  Deploy"
	@echo "    make install-service    Exibe instrução para registrar o serviço no SCM"
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
# Proto
# -----------------------------------------------------------------------------

# Gera código Go a partir do schema Protobuf.
# Requer: protoc (https://github.com/protocolbuffers/protobuf/releases)
#         protoc-gen-go (go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
proto:
	protoc \
		--go_out=. \
		--go_opt=module=github.com/mathrmm/watchdog-monitor \
		proto/metrics.proto

# -----------------------------------------------------------------------------
# Testes
# -----------------------------------------------------------------------------
test:
	go test ./...

# -----------------------------------------------------------------------------
# Execução local
# -----------------------------------------------------------------------------

# Compila e executa o binário Windows localmente.
# Requer WSL2 com integração Windows habilitada (ou Windows Terminal).
# O watchdog.toml é copiado para bin/ automaticamente.
# NATS deve estar acessível na URL configurada em watchdog.toml.
#
# Exemplo de uso com versão específica:
#   make run VERSION=1.0.0
run: build
	@[ -f bin/watchdog.toml ] || cp watchdog.toml bin/watchdog.toml
	@echo ""
	@echo "  Iniciando Watchdog Monitor v$(VERSION)"
	@echo "  Config: bin/watchdog.toml"
	@echo "  Log:    bin/watchdog.log  (use 'make logs' em outro terminal)"
	@echo "  Ctrl+C para encerrar"
	@echo ""
	./bin/watchdog.exe

stop:
	taskkill.exe /f /im watchdog.ex

# Monitora o arquivo de log em tempo real.
# Execute em um terminal separado enquanto 'make run' está ativo.
logs:
	@tail -f bin/watchdog.log 2>/dev/null || \
		(echo "Log não encontrado em bin/watchdog.log — execute 'make run' primeiro." && exit 1)

# -----------------------------------------------------------------------------
# Deploy / Install
# -----------------------------------------------------------------------------

# Exibe o comando sc create para registrar o serviço no Windows SCM.
# NÃO executa automaticamente — copie e execute no terminal Windows com privilégios de Administrador.
# Use uma conta de serviço dedicada com privilégios mínimos (SEC02).
#
# Permissões mínimas necessárias para a conta de serviço (SEC01, SEC02):
#   - Leitura do diretório de instalação (watchdog.toml, watchdog.exe)
#   - Escrita no diretório de log (watchdog.log, rotação lumberjack)
#   - Acesso WMI: namespace root\cimv2 (GPU, processos) e root\wmi (temperatura)
#   - Recomendado: conta "Local Service" ou conta dedicada, não SYSTEM
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
