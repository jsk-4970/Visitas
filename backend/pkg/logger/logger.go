package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel string

const (
	LogLevelDebug LogLevel = "DEBUG"
	LogLevelInfo  LogLevel = "INFO"
	LogLevelWarn  LogLevel = "WARN"
	LogLevelError LogLevel = "ERROR"
	LogLevelFatal LogLevel = "FATAL"
)

// Logger provides structured logging functionality
type Logger struct {
	level  LogLevel
	output io.Writer
}

// LogEntry represents a structured log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	TraceID   string                 `json:"trace_id,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// New creates a new Logger instance
func New(level LogLevel) *Logger {
	return &Logger{
		level:  level,
		output: os.Stdout,
	}
}

// NewWithOutput creates a new Logger with custom output
func NewWithOutput(level LogLevel, output io.Writer) *Logger {
	return &Logger{
		level:  level,
		output: output,
	}
}

// SetLevel sets the log level
func (l *Logger) SetLevel(level LogLevel) {
	l.level = level
}

// Debug logs a debug message
func (l *Logger) Debug(msg string, fields ...map[string]interface{}) {
	l.log(LogLevelDebug, msg, nil, fields...)
}

// DebugContext logs a debug message with context
func (l *Logger) DebugContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.logContext(ctx, LogLevelDebug, msg, nil, fields...)
}

// Info logs an info message
func (l *Logger) Info(msg string, fields ...map[string]interface{}) {
	l.log(LogLevelInfo, msg, nil, fields...)
}

// InfoContext logs an info message with context
func (l *Logger) InfoContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.logContext(ctx, LogLevelInfo, msg, nil, fields...)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string, fields ...map[string]interface{}) {
	l.log(LogLevelWarn, msg, nil, fields...)
}

// WarnContext logs a warning message with context
func (l *Logger) WarnContext(ctx context.Context, msg string, fields ...map[string]interface{}) {
	l.logContext(ctx, LogLevelWarn, msg, nil, fields...)
}

// Error logs an error message
func (l *Logger) Error(msg string, err error, fields ...map[string]interface{}) {
	l.log(LogLevelError, msg, err, fields...)
}

// ErrorContext logs an error message with context
func (l *Logger) ErrorContext(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	l.logContext(ctx, LogLevelError, msg, err, fields...)
}

// Fatal logs a fatal message and exits the program
func (l *Logger) Fatal(msg string, err error, fields ...map[string]interface{}) {
	l.log(LogLevelFatal, msg, err, fields...)
	os.Exit(1)
}

// FatalContext logs a fatal message with context and exits the program
func (l *Logger) FatalContext(ctx context.Context, msg string, err error, fields ...map[string]interface{}) {
	l.logContext(ctx, LogLevelFatal, msg, err, fields...)
	os.Exit(1)
}

// log writes a log entry
func (l *Logger) log(level LogLevel, msg string, err error, fields ...map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   msg,
	}

	if len(fields) > 0 {
		entry.Fields = fields[0]
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.writeEntry(entry)
}

// logContext writes a log entry with context
func (l *Logger) logContext(ctx context.Context, level LogLevel, msg string, err error, fields ...map[string]interface{}) {
	if !l.shouldLog(level) {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Level:     string(level),
		Message:   msg,
	}

	// Extract trace ID from context if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		entry.TraceID = fmt.Sprintf("%v", traceID)
	}

	if len(fields) > 0 {
		entry.Fields = fields[0]
	}

	if err != nil {
		entry.Error = err.Error()
	}

	l.writeEntry(entry)
}

// shouldLog determines if a message should be logged based on log level
func (l *Logger) shouldLog(level LogLevel) bool {
	levelPriority := map[LogLevel]int{
		LogLevelDebug: 0,
		LogLevelInfo:  1,
		LogLevelWarn:  2,
		LogLevelError: 3,
		LogLevelFatal: 4,
	}

	return levelPriority[level] >= levelPriority[l.level]
}

// writeEntry writes a log entry to output
func (l *Logger) writeEntry(entry LogEntry) {
	jsonBytes, err := json.Marshal(entry)
	if err != nil {
		log.Printf("Failed to marshal log entry: %v", err)
		return
	}

	fmt.Fprintln(l.output, string(jsonBytes))
}

// Global logger instance
var defaultLogger = New(LogLevelInfo)

// SetGlobalLevel sets the global logger level
func SetGlobalLevel(level LogLevel) {
	defaultLogger.SetLevel(level)
}

// Debug logs a debug message using the global logger
func Debug(msg string, fields ...map[string]interface{}) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs an info message using the global logger
func Info(msg string, fields ...map[string]interface{}) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a warning message using the global logger
func Warn(msg string, fields ...map[string]interface{}) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs an error message using the global logger
func Error(msg string, err error, fields ...map[string]interface{}) {
	defaultLogger.Error(msg, err, fields...)
}

// Fatal logs a fatal message using the global logger and exits
func Fatal(msg string, err error, fields ...map[string]interface{}) {
	defaultLogger.Fatal(msg, err, fields...)
}
