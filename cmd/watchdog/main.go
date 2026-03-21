//go:build windows

package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"

	"golang.org/x/sys/windows/svc"

	"github.com/mathrmm/watchdog-monitor/internal/config"
	"github.com/mathrmm/watchdog-monitor/internal/logger"
	"github.com/mathrmm/watchdog-monitor/internal/service"
)

// Version is injected at build time via -ldflags "-X main.Version=x.y.z".
var Version = "dev"

func main() {
	exePath, err := os.Executable()
	if err != nil {
		log.Fatalf("failed to get executable path: %v", err)
	}
	exeDir := filepath.Dir(exePath)

	cfg, err := config.Load(filepath.Join(exeDir, "watchdog.toml"))
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	logPath := cfg.LogPath
	if logPath == "" {
		logPath = filepath.Join(exeDir, "watchdog.log")
	}
	logger.Setup(logPath)

	logger.Info("version=%s nats_url=%s", Version, cfg.NatsURL)

	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine session type: %v", err)
	}

	handler := service.NewHandler()

	if isService {
		if err := svc.Run(service.ServiceName, handler); err != nil {
			log.Fatalf("service failed: %v", err)
		}
	} else {
		// Interactive mode — run until Ctrl+C or SIGTERM.
		logger.Info("running in interactive mode (not as Windows Service)")
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		logger.Info("shutting down")
	}
}
