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
	"github.com/mathrmm/watchdog-monitor/internal/publisher"
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

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatalf("failed to get hostname: %v", err)
	}
	hostname = service.SanitizeHostname(hostname)

	logger.Info("version=%s hostname=%s nats_url=%s", Version, hostname, cfg.NatsURL)

	pub, err := publisher.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}
	defer pub.Close()

	cycleRunner := service.NewCycleRunner(hostname, pub)

	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine session type: %v", err)
	}

	if isService {
		handler := service.NewHandler()
		go cycleRunner.Run(handler.StopCh())
		if err := svc.Run(service.ServiceName, handler); err != nil {
			log.Fatalf("service failed: %v", err)
		}
	} else {
		// Interactive mode — run until Ctrl+C or SIGTERM.
		logger.Info("running in interactive mode (not as Windows Service)")
		stopCh := make(chan struct{})
		go cycleRunner.Run(stopCh)

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		close(stopCh)
		logger.Info("shutting down")
	}
}
