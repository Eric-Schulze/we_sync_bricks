package logger

import (
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"
)

// StructuredOutput implements LogOutput for structured logging (JSON) to any io.Writer
type StructuredOutput struct {
	writer  io.Writer
	encoder *json.Encoder
	mutex   sync.Mutex
}

// NewStructuredOutput creates a new structured output that writes to the given writer
func NewStructuredOutput(writer io.Writer) *StructuredOutput {
	return &StructuredOutput{
		writer:  writer,
		encoder: json.NewEncoder(writer),
	}
}

// Write outputs a log entry as structured JSON
func (s *StructuredOutput) Write(entry LogEntry) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// Create a structured log entry
	structuredEntry := struct {
		Timestamp string                 `json:"@timestamp"`
		Level     string                 `json:"level"`
		Message   string                 `json:"message"`
		Source    string                 `json:"source,omitempty"`
		Fields    map[string]interface{} `json:"fields,omitempty"`
		Version   string                 `json:"version"`
		Service   string                 `json:"service"`
	}{
		Timestamp: entry.Timestamp.UTC().Format(time.RFC3339Nano),
		Level:     entry.Level.String(),
		Message:   entry.Message,
		Source:    entry.Source,
		Fields:    entry.Fields,
		Version:   "1.0",
		Service:   "we_sync_bricks",
	}

	return s.encoder.Encode(structuredEntry)
}

// Close closes the structured output
func (s *StructuredOutput) Close() error {
	// If the writer implements io.Closer, close it
	if closer, ok := s.writer.(io.Closer); ok {
		return closer.Close()
	}
	return nil
}

// BufferedOutput wraps another LogOutput with buffering for better performance
type BufferedOutput struct {
	output     LogOutput
	buffer     []LogEntry
	bufferSize int
	mutex      sync.Mutex
	flushTimer *time.Timer
}

// NewBufferedOutput creates a new buffered output
func NewBufferedOutput(output LogOutput, bufferSize int, flushInterval time.Duration) *BufferedOutput {
	bo := &BufferedOutput{
		output:     output,
		buffer:     make([]LogEntry, 0, bufferSize),
		bufferSize: bufferSize,
	}

	// Set up periodic flush
	bo.flushTimer = time.AfterFunc(flushInterval, bo.periodicFlush)

	return bo
}

// Write adds a log entry to the buffer
func (b *BufferedOutput) Write(entry LogEntry) error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	b.buffer = append(b.buffer, entry)

	// Flush if buffer is full
	if len(b.buffer) >= b.bufferSize {
		return b.flushLocked()
	}

	return nil
}

// Flush writes all buffered entries to the underlying output
func (b *BufferedOutput) Flush() error {
	b.mutex.Lock()
	defer b.mutex.Unlock()

	return b.flushLocked()
}

// flushLocked flushes the buffer without acquiring the mutex (must be called with mutex held)
func (b *BufferedOutput) flushLocked() error {
	if len(b.buffer) == 0 {
		return nil
	}

	var lastErr error
	for _, entry := range b.buffer {
		if err := b.output.Write(entry); err != nil {
			lastErr = err
		}
	}

	// Clear the buffer
	b.buffer = b.buffer[:0]

	return lastErr
}

// periodicFlush is called by the timer to flush the buffer periodically
func (b *BufferedOutput) periodicFlush() {
	b.Flush()
	// Reset the timer
	b.flushTimer.Reset(5 * time.Second) // Could be configurable
}

// Close flushes the buffer and closes the underlying output
func (b *BufferedOutput) Close() error {
	// Stop the flush timer
	if b.flushTimer != nil {
		b.flushTimer.Stop()
	}

	// Flush any remaining entries
	if err := b.Flush(); err != nil {
		fmt.Printf("Error flushing buffer on close: %v\n", err)
	}

	// Close the underlying output
	return b.output.Close()
}

// MultiOutput allows writing to multiple outputs simultaneously
type MultiOutput struct {
	outputs []LogOutput
	mutex   sync.RWMutex
}

// NewMultiOutput creates a new multi-output that writes to all provided outputs
func NewMultiOutput(outputs ...LogOutput) *MultiOutput {
	return &MultiOutput{
		outputs: outputs,
	}
}

// AddOutput adds a new output to the multi-output
func (m *MultiOutput) AddOutput(output LogOutput) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	m.outputs = append(m.outputs, output)
}

// Write writes the log entry to all outputs
func (m *MultiOutput) Write(entry LogEntry) error {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var lastErr error
	for _, output := range m.outputs {
		if err := output.Write(entry); err != nil {
			lastErr = err
			// Continue writing to other outputs even if one fails
		}
	}

	return lastErr
}

// Close closes all outputs
func (m *MultiOutput) Close() error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var lastErr error
	for _, output := range m.outputs {
		if err := output.Close(); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
