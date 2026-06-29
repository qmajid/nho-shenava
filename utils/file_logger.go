package utils

import (
	"fmt"
	"os"
	"sync"
	"time"
)

// FileLogger writes logs to a file
type FileLogger struct {
	level Level
	file  *os.File
	mu    sync.Mutex
}

// NewFileLogger creates a new file logger
func NewFileLogger(level Level, filename string) (Logger, error) {
	f, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, err
	}

	return &FileLogger{
		level: level,
		file:  f,
	}, nil
}

// Debug logs a debug message
func (l *FileLogger) Debug(msg string) {
	l.log(DebugLevel, msg)
}

// Info logs an info message
func (l *FileLogger) Info(msg string) {
	l.log(InfoLevel, msg)
}

// Warn logs a warning message
func (l *FileLogger) Warn(msg string) {
	l.log(WarnLevel, msg)
}

// Error logs an error message
func (l *FileLogger) Error(msg string) {
	l.log(ErrorLevel, msg)
}

// SetLevel sets the log level
func (l *FileLogger) SetLevel(level Level) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.level = level
}

// GetLevel returns the current log level
func (l *FileLogger) GetLevel() Level {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.level
}

func (l *FileLogger) log(level Level, msg string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if level < l.level {
		return
	}

	prefix := LevelMap[level]
	timestamp := time.Now().Format("2006/01/02 15:04:05")
	fmt.Fprintf(l.file, "[%s] [%s] %s\n", timestamp, prefix, msg)
}

// Close closes the file logger
func (l *FileLogger) Close() {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.file != nil {
		l.file.Close()
		l.file = nil
	}
}
