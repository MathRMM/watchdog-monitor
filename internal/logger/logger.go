package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"

	"gopkg.in/natefinch/lumberjack.v2"
)

var (
	mu  sync.Mutex
	std *log.Logger
)

// Setup initializes the logger with file rotation via lumberjack.
// Logs are written to both the file and stdout.
// If logPath is empty, defaults to "watchdog.log".
func Setup(logPath string) {
	if logPath == "" {
		logPath = "watchdog.log"
	}

	rotator := &lumberjack.Logger{
		Filename:   logPath,
		MaxSize:    10, // MB before rotation
		MaxBackups: 3,
		Compress:   false,
	}

	multi := io.MultiWriter(os.Stdout, rotator)

	mu.Lock()
	defer mu.Unlock()
	std = log.New(multi, "", log.LstdFlags)
}

// Info logs an informational message.
func Info(format string, args ...any) {
	mu.Lock()
	l := std
	mu.Unlock()

	msg := fmt.Sprintf("[INFO] "+format, args...)
	if l == nil {
		log.Print(msg)
		return
	}
	l.Print(msg)
}

// Error logs an error message.
func Error(format string, args ...any) {
	mu.Lock()
	l := std
	mu.Unlock()

	msg := fmt.Sprintf("[ERROR] "+format, args...)
	if l == nil {
		log.Print(msg)
		return
	}
	l.Print(msg)
}
