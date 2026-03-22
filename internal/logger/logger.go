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
	mu         sync.Mutex
	std        *log.Logger
	lastErrMsg string // last error message logged; empty means "no suppression active"
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

	SetupWriter(io.MultiWriter(os.Stdout, rotator))
}

// SetupWriter configures the logger to write to w.
// Passing nil resets the logger to the default log.Default() behaviour.
// Intended for use in tests — production code calls Setup().
func SetupWriter(w io.Writer) {
	mu.Lock()
	defer mu.Unlock()
	lastErrMsg = "" // reset dedup state on every setup
	if w == nil {
		std = nil
		return
	}
	std = log.New(w, "", log.LstdFlags)
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

// Error logs an error message with deduplication: if the same message is logged
// consecutively it is suppressed after the first occurrence. A different message
// clears the suppression and is logged normally.
func Error(format string, args ...any) {
	mu.Lock()
	l := std
	msg := fmt.Sprintf("[ERROR] "+format, args...)
	if msg == lastErrMsg {
		mu.Unlock()
		return // suppress duplicate
	}
	lastErrMsg = msg
	mu.Unlock()

	if l == nil {
		log.Print(msg)
		return
	}
	l.Print(msg)
}
