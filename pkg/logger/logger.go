package logger

import (
	"log"
	"os"
)

type Logger struct {
	info  *log.Logger
	error *log.Logger
	debug *log.Logger
}

// NewLogger creates a new logger
func NewLogger() *Logger {
	return &Logger{
		info:  log.New(os.Stdout, "[INFO] ", log.LstdFlags),
		error: log.New(os.Stderr, "[ERROR] ", log.LstdFlags),
		debug: log.New(os.Stdout, "[DEBUG] ", log.LstdFlags),
	}
}

// Info logs an info message
func (l *Logger) Info(msg string, args ...interface{}) {
	l.info.Printf(msg, args...)
}

// Error logs an error message
func (l *Logger) Error(msg string, args ...interface{}) {
	l.error.Printf(msg, args...)
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, args ...interface{}) {
	l.debug.Printf(msg, args...)
}
