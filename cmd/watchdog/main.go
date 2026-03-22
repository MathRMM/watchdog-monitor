//go:build windows

package main

import (
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"sync"
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

	// RF14: log version, hostname and NATS URL on startup.
	logger.Info("version=%s hostname=%s nats_url=%s", Version, hostname, cfg.NatsURL)

	pub, err := publisher.Connect(cfg.NatsURL)
	if err != nil {
		log.Fatalf("failed to connect to NATS: %v", err)
	}

	cycleRunner := service.NewCycleRunner(hostname, pub)

	// runCycle starts the cycle goroutine and returns channels to stop it and
	// wait for it to finish — enabling gracious shutdown before pub.Close().
	runCycle := func(stopCh <-chan struct{}) (done <-chan struct{}) {
		ch := make(chan struct{})
		go func() {
			defer close(ch)
			cycleRunner.Run(stopCh)
		}()
		return ch
	}

	isService, err := svc.IsWindowsService()
	if err != nil {
		log.Fatalf("failed to determine session type: %v", err)
	}

	var wg sync.WaitGroup
	wg.Add(1)

	if isService {
		handler := service.NewHandler()
		done := runCycle(handler.StopCh())
		go func() {
			defer wg.Done()
			<-done
		}()
		if err := svc.Run(service.ServiceName, handler); err != nil {
			log.Fatalf("service failed: %v", err)
		}
	} else {
		// Interactive mode — run until Ctrl+C or SIGTERM.
		logger.Info("running in interactive mode (not as Windows Service)")
		stopCh := make(chan struct{})
		done := runCycle(stopCh)
		go func() {
			defer wg.Done()
			<-done
		}()

		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		close(stopCh)
	}

	// Wait for the cycle goroutine to finish before closing the publisher.
	// This ensures no in-flight Publish() call races with pub.Close() (RNF04).
	wg.Wait()
	logger.Info("shutting down")
	pub.Close()
}
