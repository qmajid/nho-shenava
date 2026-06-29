package utils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// TerminalLogger prints logs to terminal
type TerminalLogger struct {
	level Level
	mu    sync.Mutex
}

// NewTerminalLogger creates a new terminal logger
func NewTerminalLogger(level Level) *TerminalLogger {
	return &TerminalLogger{
		level: level,
	}
}

// Debug logs a debug message
func (l *TerminalLogger) Debug(msg string) {
	l.log(DebugLevel, msg)
}

// Info logs an info message
func (l *TerminalLogger) Info(msg string) {
	l.log(InfoLevel, msg)
}

// Warn logs a warning message
func (l *TerminalLogger) Warn(msg string) {
	l.log(WarnLevel, msg)
}

// Error logs an error message
func (l *TerminalLogger) Error(msg string) {
	l.log(ErrorLevel, msg)
}

// SetLevel sets the log level
func (l *TerminalLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *TerminalLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *TerminalLogger) log(level Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	prefix := LevelMap[level]
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(os.Stdout, "[%s] [%s] %s\n", timestamp, prefix, msg)
}
