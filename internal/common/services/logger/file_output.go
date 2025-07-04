package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"
)

// FileOutput implements LogOutput for file-based logging
type FileOutput struct {
	file     *os.File
	encoder  *json.Encoder
	mutex    sync.Mutex
	filePath string
	format   FileFormat
}

// FileFormat determines how log entries are written to the file
type FileFormat int

const (
	FormatJSON FileFormat = iota
	FormatText
)

// FileOutputConfig contains configuration for file output
type FileOutputConfig struct {
	FilePath   string
	Format     FileFormat
	MaxSize    int64 // Maximum file size in bytes (0 = no limit)
	MaxAge     int   // Maximum age in days (0 = no limit)
	MaxBackups int   // Maximum number of backup files (0 = no limit)
}

// NewFileOutput creates a new file output with JSON format
func NewFileOutput(filePath string) (*FileOutput, error) {
	return NewFileOutputWithConfig(FileOutputConfig{
		FilePath: filePath,
		Format:   FormatJSON,
	})
}

// NewFileOutputWithConfig creates a new file output with custom configuration
func NewFileOutputWithConfig(config FileOutputConfig) (*FileOutput, error) {
	// Ensure directory exists
	dir := filepath.Dir(config.FilePath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}
	
	// Open or create the log file
	file, err := os.OpenFile(config.FilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}
	
	output := &FileOutput{
		file:     file,
		filePath: config.FilePath,
		format:   config.Format,
	}
	
	if config.Format == FormatJSON {
		output.encoder = json.NewEncoder(file)
	}
	
	return output, nil
}

// Write outputs a log entry to the file
func (f *FileOutput) Write(entry LogEntry) error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	switch f.format {
	case FormatJSON:
		return f.writeJSON(entry)
	case FormatText:
		return f.writeText(entry)
	default:
		return f.writeJSON(entry)
	}
}

// writeJSON writes the log entry as JSON
func (f *FileOutput) writeJSON(entry LogEntry) error {
	jsonEntry := struct {
		Timestamp string                 `json:"timestamp"`
		Level     string                 `json:"level"`
		Message   string                 `json:"message"`
		Source    string                 `json:"source,omitempty"`
		Fields    map[string]interface{} `json:"fields,omitempty"`
	}{
		Timestamp: entry.Timestamp.UTC().Format(time.RFC3339),
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Source:    entry.Source,
		Fields:    entry.Fields,
	}
	
	return f.encoder.Encode(jsonEntry)
}

// writeText writes the log entry as formatted text
func (f *FileOutput) writeText(entry LogEntry) error {
	timestamp := entry.Timestamp.Format("2006-01-02 15:04:05")
	
	logLine := fmt.Sprintf("[%s] %s %s", timestamp, entry.Level.String(), entry.Message)
	
	// Add fields if any
	if len(entry.Fields) > 0 {
		for key, value := range entry.Fields {
			logLine += fmt.Sprintf(" %s=%v", key, value)
		}
	}
	
	logLine += "\n"
	
	_, err := f.file.WriteString(logLine)
	if err != nil {
		return err
	}
	
	// Ensure data is written to disk
	return f.file.Sync()
}

// Close closes the file output
func (f *FileOutput) Close() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if f.file != nil {
		err := f.file.Close()
		f.file = nil
		return err
	}
	return nil
}

// RotateFile rotates the current log file by renaming it and creating a new one
func (f *FileOutput) RotateFile() error {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	// Close current file
	if f.file != nil {
		f.file.Close()
	}
	
	// Rename current file with timestamp
	timestamp := time.Now().Format("20060102_150405")
	rotatedPath := fmt.Sprintf("%s.%s", f.filePath, timestamp)
	
	if err := os.Rename(f.filePath, rotatedPath); err != nil {
		// If rename fails, just continue with new file
		fmt.Fprintf(os.Stderr, "Failed to rotate log file: %v\n", err)
	}
	
	// Create new file
	file, err := os.OpenFile(f.filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return fmt.Errorf("failed to create new log file: %w", err)
	}
	
	f.file = file
	if f.format == FormatJSON {
		f.encoder = json.NewEncoder(file)
	}
	
	return nil
}

// GetFileSize returns the current size of the log file
func (f *FileOutput) GetFileSize() (int64, error) {
	f.mutex.Lock()
	defer f.mutex.Unlock()
	
	if f.file == nil {
		return 0, fmt.Errorf("file is not open")
	}
	
	stat, err := f.file.Stat()
	if err != nil {
		return 0, err
	}
	
	return stat.Size(), nil
}