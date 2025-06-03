package logger

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"
)

// Logger provides structured logging capabilities
type Logger struct {
	name string
}

// LogLevel represents different logging levels
type LogLevel string

const (
	LevelDebug LogLevel = "DEBUG"
	LevelInfo  LogLevel = "INFO"
	LevelWarn  LogLevel = "WARN"
	LevelError LogLevel = "ERROR"
)

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp time.Time              `json:"timestamp"`
	Level     LogLevel               `json:"level"`
	Logger    string                 `json:"logger"`
	Message   string                 `json:"message"`
	Data      map[string]interface{} `json:"data,omitempty"`
}

// New creates a new logger instance with the given name
func New(name string) *Logger {
	return &Logger{
		name: name,
	}
}

// Debug logs a debug message with optional data
func (l *Logger) Debug(message string, data ...map[string]interface{}) {
	l.log(LevelDebug, message, data...)
}

// Info logs an info message with optional data
func (l *Logger) Info(message string, data ...map[string]interface{}) {
	l.log(LevelInfo, message, data...)
}

// Warn logs a warning message with optional data
func (l *Logger) Warn(message string, data ...map[string]interface{}) {
	l.log(LevelWarn, message, data...)
}

// Error logs an error message with optional data
func (l *Logger) Error(message string, data ...map[string]interface{}) {
	l.log(LevelError, message, data...)
}

// log is the internal logging method
func (l *Logger) log(level LogLevel, message string, data ...map[string]interface{}) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Logger:    l.name,
		Message:   message,
	}

	// Merge all data maps if provided
	if len(data) > 0 {
		entry.Data = make(map[string]interface{})
		for _, d := range data {
			for k, v := range d {
				entry.Data[k] = v
			}
		}
	}

	// Check if we should use structured logging
	if os.Getenv("LOG_FORMAT") == "json" {
		// JSON structured logging
		jsonBytes, err := json.Marshal(entry)
		if err != nil {
			// Fallback to simple logging if JSON marshal fails
			log.Printf("[%s] %s: %s", level, l.name, message)
			return
		}
		fmt.Println(string(jsonBytes))
	} else {
		// Simple text logging
		if len(data) > 0 {
			dataStr, _ := json.Marshal(entry.Data)
			log.Printf("[%s] %s: %s - %s", level, l.name, message, string(dataStr))
		} else {
			log.Printf("[%s] %s: %s", level, l.name, message)
		}
	}
}

// WithField creates a new logger with a persistent field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	// For simplicity, we'll just return the same logger
	// In a more sophisticated implementation, this would create a new logger
	// with persistent fields
	return l
}

// WithFields creates a new logger with persistent fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	// For simplicity, we'll just return the same logger
	// In a more sophisticated implementation, this would create a new logger
	// with persistent fields
	return l
}