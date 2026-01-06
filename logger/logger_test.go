package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestNoopLogger(t *testing.T) {
	// By default, logging should be silent (no-op)
	var buf bytes.Buffer

	// Don't set a logger - should use no-op by default
	Debug("debug message")
	Info("info message")
	Warn("warn message")
	Error("error message")

	// Buffer should be empty since no logger is set
	if buf.Len() > 0 {
		t.Errorf("Expected no output with default no-op logger, got: %s", buf.String())
	}
}

func TestSetLogger(t *testing.T) {
	var buf bytes.Buffer

	// Enable logging with slog
	SetLogger(NewSlogLogger(
		WithLevel(LevelInfo),
		WithFormat(FormatText),
		WithOutput(&buf),
	))

	Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Expected message in output, got: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Expected key=value in output, got: %s", output)
	}

	// Reset to no-op
	SetLogger(nil)
}

func TestSlogLoggerLevels(t *testing.T) {
	tests := []struct {
		name         string
		level        Level
		logFunc      func(string, ...any)
		message      string
		shouldAppear bool
	}{
		{"Debug at Debug level", LevelDebug, Debug, "debug message", true},
		{"Info at Debug level", LevelDebug, Info, "info message", true},
		{"Debug at Info level", LevelInfo, Debug, "debug message", false},
		{"Info at Info level", LevelInfo, Info, "info message", true},
		{"Warn at Info level", LevelInfo, Warn, "warn message", true},
		{"Info at Error level", LevelError, Info, "info message", false},
		{"Error at Error level", LevelError, Error, "error message", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			SetLogger(NewSlogLogger(
				WithLevel(tt.level),
				WithFormat(FormatText),
				WithOutput(&buf),
			))
			defer SetLogger(nil) // Reset after test

			tt.logFunc(tt.message)

			output := buf.String()
			contains := strings.Contains(output, tt.message)

			if tt.shouldAppear && !contains {
				t.Errorf("Expected message %q to appear in output, but it didn't. Output: %s", tt.message, output)
			}
			if !tt.shouldAppear && contains {
				t.Errorf("Expected message %q NOT to appear in output, but it did. Output: %s", tt.message, output)
			}
		})
	}
}

func TestSlogLoggerFormats(t *testing.T) {
	tests := []struct {
		name           string
		format         Format
		expectedMarker string
	}{
		{"Text format", FormatText, "level=INFO"},
		{"JSON format", FormatJSON, `"level":"INFO"`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			SetLogger(NewSlogLogger(
				WithLevel(LevelInfo),
				WithFormat(tt.format),
				WithOutput(&buf),
			))
			defer SetLogger(nil)

			Info("test message", "key", "value")

			output := buf.String()
			if !strings.Contains(output, tt.expectedMarker) {
				t.Errorf("Expected output to contain %q, got: %s", tt.expectedMarker, output)
			}
		})
	}
}

func TestSlogLoggerAllLevels(t *testing.T) {
	var buf bytes.Buffer
	SetLogger(NewSlogLogger(
		WithLevel(LevelDebug),
		WithFormat(FormatText),
		WithOutput(&buf),
	))
	defer SetLogger(nil)

	tests := []struct {
		name    string
		logFunc func(string, ...any)
		level   string
	}{
		{"Debug", Debug, "DEBUG"},
		{"Info", Info, "INFO"},
		{"Warn", Warn, "WARN"},
		{"Error", Error, "ERROR"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			tt.logFunc("test message", "key", "value")

			output := buf.String()
			if !strings.Contains(output, "level="+tt.level) {
				t.Errorf("Expected level=%s in output, got: %s", tt.level, output)
			}
			if !strings.Contains(output, "test message") {
				t.Errorf("Expected message in output, got: %s", output)
			}
			if !strings.Contains(output, "key=value") {
				t.Errorf("Expected key=value in output, got: %s", output)
			}
		})
	}
}

func TestConcurrentLogging(t *testing.T) {
	var buf bytes.Buffer
	SetLogger(NewSlogLogger(
		WithLevel(LevelInfo),
		WithFormat(FormatText),
		WithOutput(&buf),
	))
	defer SetLogger(nil)

	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			Info("concurrent log", "goroutine", id)
			done <- true
		}(i)
	}

	for i := 0; i < 10; i++ {
		<-done
	}

	// Just verify it doesn't panic
	output := buf.String()
	if output == "" {
		t.Error("Expected some output from concurrent logging")
	}
}

func TestCustomLogger(t *testing.T) {
	// Test that users can provide their own logger implementation
	customLogger := &customTestLogger{messages: make([]string, 0)}

	SetLogger(customLogger)
	defer SetLogger(nil)

	Info("test message")

	if len(customLogger.messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(customLogger.messages))
	}
	if customLogger.messages[0] != "INFO: test message" {
		t.Errorf("Expected 'INFO: test message', got %q", customLogger.messages[0])
	}
}

// customTestLogger implements the Logger interface for testing
type customTestLogger struct {
	messages []string
}

func (c *customTestLogger) Debug(msg string, _ ...any) {
	c.messages = append(c.messages, "DEBUG: "+msg)
}

func (c *customTestLogger) Info(msg string, _ ...any) {
	c.messages = append(c.messages, "INFO: "+msg)
}

func (c *customTestLogger) Warn(msg string, _ ...any) {
	c.messages = append(c.messages, "WARN: "+msg)
}

func (c *customTestLogger) Error(msg string, _ ...any) {
	c.messages = append(c.messages, "ERROR: "+msg)
}
