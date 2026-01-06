// Package logger provides an optional structured logging interface for findata-go.
//
// By default, logging is DISABLED (silent). Users can opt-in to logging by calling SetLogger
// with their preferred logger implementation.
//
// The package provides a ready-to-use slog-based implementation via NewSlogLogger() for convenience.
//
// Example - Enable logging with slog:
//
//	import "github.com/Vikramarjuna/findata-go/logger"
//
//	// Enable logging with default settings (INFO level, text format, stderr)
//	logger.SetLogger(logger.NewSlogLogger())
//
//	// Or with custom configuration
//	logger.SetLogger(logger.NewSlogLogger(
//	    logger.WithLevel(logger.LevelDebug),
//	    logger.WithFormat(logger.FormatJSON),
//	))
//
// Example - Use your own logger:
//
//	type MyLogger struct{}
//	func (l *MyLogger) Debug(msg string, keysAndValues ...any) { /* ... */ }
//	func (l *MyLogger) Info(msg string, keysAndValues ...any) { /* ... */ }
//	func (l *MyLogger) Warn(msg string, keysAndValues ...any) { /* ... */ }
//	func (l *MyLogger) Error(msg string, keysAndValues ...any) { /* ... */ }
//
//	logger.SetLogger(&MyLogger{})
package logger

import (
	"io"
	"log/slog"
	"os"
	"sync"
)

// Logger is the interface for logging in findata-go.
// Users can provide their own implementation or use the provided slog-based logger.
type Logger interface {
	// Debug logs a debug message with optional key-value pairs
	Debug(msg string, keysAndValues ...any)
	// Info logs an info message with optional key-value pairs
	Info(msg string, keysAndValues ...any)
	// Warn logs a warning message with optional key-value pairs
	Warn(msg string, keysAndValues ...any)
	// Error logs an error message with optional key-value pairs
	Error(msg string, keysAndValues ...any)
}

// Level represents the severity of a log message
type Level int

const (
	// LevelDebug is the debug log level
	LevelDebug Level = iota
	// LevelInfo is the info log level
	LevelInfo
	// LevelWarn is the warning log level
	LevelWarn
	// LevelError is the error log level
	LevelError
)

// Format represents the output format for logs
type Format int

const (
	// FormatText outputs logs in human-readable text format
	FormatText Format = iota
	// FormatJSON outputs logs in JSON format
	FormatJSON
)

var (
	globalLogger Logger = &noopLogger{} // Default: silent
	mu           sync.RWMutex
)

// noopLogger is a logger that does nothing (silent by default)
type noopLogger struct{}

// Debug does nothing (no-op implementation).
func (n *noopLogger) Debug(_ string, _ ...any) {}

// Info does nothing (no-op implementation).
func (n *noopLogger) Info(_ string, _ ...any) {}

// Warn does nothing (no-op implementation).
func (n *noopLogger) Warn(_ string, _ ...any) {}

// Error does nothing (no-op implementation).
func (n *noopLogger) Error(_ string, _ ...any) {}

// SetLogger sets the global logger for findata-go.
// By default, logging is disabled. Call this to enable logging.
//
// Example:
//
//	logger.SetLogger(logger.NewSlogLogger()) // Enable with defaults
func SetLogger(l Logger) {
	mu.Lock()
	defer mu.Unlock()
	if l == nil {
		globalLogger = &noopLogger{}
	} else {
		globalLogger = l
	}
}

// Debug logs a debug message with optional key-value pairs
func Debug(msg string, keysAndValues ...any) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Debug(msg, keysAndValues...)
}

// Info logs an info message with optional key-value pairs
func Info(msg string, keysAndValues ...any) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Info(msg, keysAndValues...)
}

// Warn logs a warning message with optional key-value pairs
func Warn(msg string, keysAndValues ...any) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Warn(msg, keysAndValues...)
}

// Error logs an error message with optional key-value pairs
func Error(msg string, keysAndValues ...any) {
	mu.RLock()
	defer mu.RUnlock()
	globalLogger.Error(msg, keysAndValues...)
}

// slogLogger is a Logger implementation using Go's standard slog package
type slogLogger struct {
	logger *slog.Logger
}

// Debug logs a debug message with optional key-value pairs.
func (s *slogLogger) Debug(msg string, keysAndValues ...any) {
	s.logger.Debug(msg, keysAndValues...)
}

// Info logs an info message with optional key-value pairs.
func (s *slogLogger) Info(msg string, keysAndValues ...any) {
	s.logger.Info(msg, keysAndValues...)
}

// Warn logs a warning message with optional key-value pairs.
func (s *slogLogger) Warn(msg string, keysAndValues ...any) {
	s.logger.Warn(msg, keysAndValues...)
}

func (s *slogLogger) Error(msg string, keysAndValues ...any) {
	s.logger.Error(msg, keysAndValues...)
}

// SlogOption is a configuration option for NewSlogLogger
type SlogOption func(*slogConfig)

type slogConfig struct {
	level  Level
	format Format
	output io.Writer
}

// WithLevel sets the log level for the slog logger
func WithLevel(level Level) SlogOption {
	return func(c *slogConfig) {
		c.level = level
	}
}

// WithFormat sets the output format for the slog logger
func WithFormat(format Format) SlogOption {
	return func(c *slogConfig) {
		c.format = format
	}
}

// WithOutput sets the output writer for the slog logger
func WithOutput(output io.Writer) SlogOption {
	return func(c *slogConfig) {
		c.output = output
	}
}

// NewSlogLogger creates a new Logger using Go's standard slog package.
// By default, it logs at INFO level in text format to stderr.
//
// Example:
//
//	// Default configuration
//	logger.SetLogger(logger.NewSlogLogger())
//
//	// Custom configuration
//	logger.SetLogger(logger.NewSlogLogger(
//	    logger.WithLevel(logger.LevelDebug),
//	    logger.WithFormat(logger.FormatJSON),
//	    logger.WithOutput(logFile),
//	))
func NewSlogLogger(opts ...SlogOption) Logger {
	config := &slogConfig{
		level:  LevelInfo,
		format: FormatText,
		output: os.Stderr,
	}

	for _, opt := range opts {
		opt(config)
	}

	var slogLevel slog.Level
	switch config.level {
	case LevelDebug:
		slogLevel = slog.LevelDebug
	case LevelInfo:
		slogLevel = slog.LevelInfo
	case LevelWarn:
		slogLevel = slog.LevelWarn
	case LevelError:
		slogLevel = slog.LevelError
	default:
		slogLevel = slog.LevelInfo
	}

	handlerOpts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var handler slog.Handler
	switch config.format {
	case FormatJSON:
		handler = slog.NewJSONHandler(config.output, handlerOpts)
	default:
		handler = slog.NewTextHandler(config.output, handlerOpts)
	}

	return &slogLogger{
		logger: slog.New(handler),
	}
}
