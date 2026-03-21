//go:build windows

package service

import (
	"github.com/mathrmm/watchdog-monitor/internal/logger"
	"golang.org/x/sys/windows/svc"
)

// ServiceName is the name registered in the Windows SCM.
const ServiceName = "WatchdogMonitor"

// Handler implements the golang.org/x/sys/windows/svc.Handler interface.
type Handler struct {
	stopCh chan struct{}
}

// NewHandler creates a new service Handler.
func NewHandler() *Handler {
	return &Handler{
		stopCh: make(chan struct{}),
	}
}

// Execute is called by the Windows SCM when the service starts.
// It signals running state, waits for Stop/Shutdown from SCM, then returns.
func (h *Handler) Execute(args []string, r <-chan svc.ChangeRequest, s chan<- svc.Status) (bool, uint32) {
	s <- svc.Status{State: svc.StartPending}

	logger.Info("Watchdog Monitor service starting")

	s <- svc.Status{
		State:   svc.Running,
		Accepts: svc.AcceptStop | svc.AcceptShutdown,
	}

loop:
	for {
		select {
		case c := <-r:
			switch c.Cmd {
			case svc.Stop, svc.Shutdown:
				break loop
			}
		}
	}

	logger.Info("Watchdog Monitor service stopping")
	s <- svc.Status{State: svc.StopPending}
	close(h.stopCh)
	return false, 0
}

// StopCh returns a channel that is closed when the SCM signals stop.
func (h *Handler) StopCh() <-chan struct{} {
	return h.stopCh
}
