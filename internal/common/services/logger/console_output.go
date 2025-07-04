package logger

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

// ConsoleOutput implements LogOutput for console/terminal output
type ConsoleOutput struct {
	colorEnabled bool
}

// NewConsoleOutput creates a new console output
func NewConsoleOutput() *ConsoleOutput {
	return &ConsoleOutput{
		colorEnabled: true, // Could be configurable based on terminal capabilities
	}
}

// NewConsoleOutputNoColor creates a new console output without colors
func NewConsoleOutputNoColor() *ConsoleOutput {
	return &ConsoleOutput{
		colorEnabled: false,
	}
}

// Write outputs a log entry to the console
func (c *ConsoleOutput) Write(entry LogEntry) error {
	var levelStr string
	if c.colorEnabled {
		levelStr = c.colorizeLevel(entry.Level)
	} else {
		levelStr = entry.Level.String()
	}
	
	// Format timestamp
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	
	// Build the main log line
	logLine := fmt.Sprintf("[%s] %s %s", timestamp, levelStr, entry.Message)
	
	// Add fields if any
	if len(entry.Fields) > 0 {
		fieldsStr := c.formatFields(entry.Fields)
		logLine += " " + fieldsStr
	}
	
	// Write to appropriate stream (stderr for errors, stdout for others)
	var output *os.File
	if entry.Level >= LogError {
		output = os.Stderr
	} else {
		output = os.Stdout
	}
	
	_, err := fmt.Fprintln(output, logLine)
	return err
}

// Close closes the console output (no-op for console)
func (c *ConsoleOutput) Close() error {
	return nil
}

// colorizeLevel adds ANSI color codes to the log level
func (c *ConsoleOutput) colorizeLevel(level LogLevel) string {
	if !c.colorEnabled {
		return level.String()
	}
	
	switch level {
	case LogDebug:
		return "\033[36mDEBUG\033[0m" // Cyan
	case LogInfo:
		return "\033[32mINFO\033[0m"  // Green
	case LogWarn:
		return "\033[33mWARN\033[0m"  // Yellow
	case LogError:
		return "\033[31mERROR\033[0m" // Red
	default:
		return level.String()
	}
}

// formatFields formats the log fields as key=value pairs
func (c *ConsoleOutput) formatFields(fields map[string]interface{}) string {
	if len(fields) == 0 {
		return ""
	}
	
	// Sort keys for consistent output
	keys := make([]string, 0, len(fields))
	for k := range fields {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	
	// Build key=value pairs
	pairs := make([]string, 0, len(fields))
	for _, key := range keys {
		value := fields[key]
		pairs = append(pairs, fmt.Sprintf("%s=%v", key, value))
	}
	
	return strings.Join(pairs, " ")
}