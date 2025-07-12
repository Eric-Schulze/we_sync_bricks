package logger

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"time"
)

// LogLevel represents the severity level of a log message
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// LogEntry represents a single log message
type LogEntry struct {
	Level     LogLevel
	Message   string
	Fields    map[string]interface{}
	Timestamp time.Time
	Source    string
}

// LogOutput defines the interface for different log output formats
type LogOutput interface {
	Write(entry LogEntry) error
	Close() error
}

// Logger is the main logger struct that manages multiple outputs
type Logger struct {
	outputs  []LogOutput
	minLevel LogLevel
	context  map[string]interface{}
}

// NewLogger creates a new logger instance
func NewLogger(minLevel LogLevel) *Logger {
	return &Logger{
		outputs:  make([]LogOutput, 0),
		minLevel: minLevel,
		context:  make(map[string]interface{}),
	}
}

// AddOutput adds a new output destination to the logger
func (l *Logger) AddOutput(output LogOutput) {
	l.outputs = append(l.outputs, output)
}

// WithContext adds context fields that will be included in all log messages
func (l *Logger) WithContext(key string, value interface{}) *Logger {
	newLogger := &Logger{
		outputs:  l.outputs,
		minLevel: l.minLevel,
		context:  make(map[string]interface{}),
	}

	// Copy existing context
	for k, v := range l.context {
		newLogger.context[k] = v
	}

	// Add new context field
	newLogger.context[key] = value

	return newLogger
}

// log is the internal method that handles the actual logging
func (l *Logger) log(level LogLevel, message string, fields ...interface{}) {
	if level < l.minLevel {
		return
	}

	entry := LogEntry{
		Level:     level,
		Message:   message,
		Fields:    make(map[string]interface{}),
		Timestamp: time.Now(),
		Source:    getCallerInfo(),
	}

	// Add context fields
	for k, v := range l.context {
		entry.Fields[k] = v
	}

	// Add provided fields (key-value pairs)
	for i := 0; i < len(fields); i += 2 {
		if i+1 < len(fields) {
			if key, ok := fields[i].(string); ok {
				entry.Fields[key] = fields[i+1]
			}
		}
	}

	// Write to all outputs
	for _, output := range l.outputs {
		if err := output.Write(entry); err != nil {
			// If logging fails, write to stderr as fallback
			fmt.Fprintf(os.Stderr, "Logger error: %v\n", err)
		}
	}
}

// Debug logs a debug message
func (l *Logger) Debug(message string, fields ...interface{}) {
	l.log(LogDebug, message, fields...)
}

// Info logs an info message
func (l *Logger) Info(message string, fields ...interface{}) {
	l.log(LogInfo, message, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(message string, fields ...interface{}) {
	l.log(LogWarn, message, fields...)
}

// Error logs an error message
func (l *Logger) Error(message string, fields ...interface{}) {
	l.log(LogError, message, fields...)
}

// Close closes all outputs
func (l *Logger) Close() error {
	var lastErr error
	for _, output := range l.outputs {
		if err := output.Close(); err != nil {
			lastErr = err
		}
	}
	return lastErr
}

// getCallerInfo returns information about the calling function
func getCallerInfo() string {
	// This is a simplified version - could be enhanced to provide more detailed caller info
	return "unknown"
}

// String returns the string representation of LogLevel
func (l LogLevel) String() string {
	switch l {
	case LogDebug:
		return "DEBUG"
	case LogInfo:
		return "INFO"
	case LogWarn:
		return "WARN"
	case LogError:
		return "ERROR"
	default:
		return "UNKNOWN"
	}
}

// Global logger instance
var defaultLogger *Logger

// InitializeDefaultLogger initializes the global logger with default outputs
func InitializeDefaultLogger(logLevel LogLevel, logFilePath string) error {
	defaultLogger = NewLogger(logLevel)

	// Add console output
	consoleOutput := NewConsoleOutput()
	defaultLogger.AddOutput(consoleOutput)

	// Add file output if path is provided
	if logFilePath != "" {
		// Ensure log directory exists
		if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
			return fmt.Errorf("failed to create log directory: %w", err)
		}

		fileOutput, err := NewFileOutput(logFilePath)
		if err != nil {
			return fmt.Errorf("failed to create file output: %w", err)
		}
		defaultLogger.AddOutput(fileOutput)
	}

	return nil
}

// Package-level logging functions that use the default logger
func Debug(message string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Debug(message, fields...)
	}
}

func Info(message string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Info(message, fields...)
	}
}

func Warn(message string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Warn(message, fields...)
	}
}

func Error(message string, fields ...interface{}) {
	if defaultLogger != nil {
		defaultLogger.Error(message, fields...)
	}
}

func WithContext(key string, value interface{}) *Logger {
	if defaultLogger != nil {
		return defaultLogger.WithContext(key, value)
	}
	return NewLogger(LogInfo)
}

// SlogAdapter creates an slog.Logger that uses our custom logger
func SlogAdapter() *slog.Logger {
	if defaultLogger == nil {
		return slog.Default()
	}

	handler := &slogHandler{logger: defaultLogger}
	return slog.New(handler)
}

// slogHandler implements slog.Handler to bridge with our custom logger
type slogHandler struct {
	logger *Logger
}

func (h *slogHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return true // Let our logger handle level filtering
}

func (h *slogHandler) Handle(ctx context.Context, record slog.Record) error {
	var logLevel LogLevel
	switch {
	case record.Level >= slog.LevelError:
		logLevel = LogError
	case record.Level >= slog.LevelWarn:
		logLevel = LogWarn
	case record.Level >= slog.LevelInfo:
		logLevel = LogInfo
	default:
		logLevel = LogDebug
	}

	fields := make([]interface{}, 0)
	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, attr.Key, attr.Value.Any())
		return true
	})

	h.logger.log(logLevel, record.Message, fields...)
	return nil
}

func (h *slogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newLogger := h.logger
	for _, attr := range attrs {
		newLogger = newLogger.WithContext(attr.Key, attr.Value.Any())
	}
	return &slogHandler{logger: newLogger}
}

func (h *slogHandler) WithGroup(name string) slog.Handler {
	return h // Simplified - could implement grouping if needed
}
