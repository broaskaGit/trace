package trace

import (
	"os"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// SugarLogger implements the LoggerInterface with a real zap logger
type SugarLogger struct {
	Log *zap.Logger
}

// New creates the fastest possible logger configuration
// level: minimum log level (e.g., zapcore.InfoLevel)
// prefix: logger name prefix for all messages
// logFile: optional file to write logs to (pass nil to log to stdout only)
// To disable logging completely, use zapcore.Level(127)
func New(level zapcore.Level, prefix string, logFile *os.File) Logger {
	// Fastest possible encoder config
	encoderConfig := zapcore.EncoderConfig{
		MessageKey:     "msg",
		LevelKey:       "level",
		TimeKey:        "ts",
		NameKey:        "logger",
		CallerKey:      "",                                                 // disabled for speed
		StacktraceKey:  "",                                                 // disabled for speed
		EncodeLevel:    zapcore.CapitalColorLevelEncoder,                   // colored level in caps
		EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"), // human readable time
		EncodeDuration: zapcore.StringDurationEncoder,
	}

	// Create stdout writer
	stdoutSink := zapcore.Lock(os.Stdout)

	var core zapcore.Core

	// If logFile is provided, create a multi-output core
	if logFile != nil {
		// Create file sink
		fileSink := zapcore.Lock(logFile)

		// Create a core that writes to both stdout and file
		core = zapcore.NewTee(
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), stdoutSink, level),
			zapcore.NewCore(zapcore.NewConsoleEncoder(encoderConfig), fileSink, level),
		)
	} else {
		// Standard stdout-only core
		core = zapcore.NewCore(
			zapcore.NewConsoleEncoder(encoderConfig),
			stdoutSink,
			level,
		)
	}

	if prefix != "" {
		return &SugarLogger{
			Log: zap.New(core).Named(prefix),
		}
	}

	// Build the logger with minimal options for speed
	return &SugarLogger{
		Log: zap.New(core),
	}
}

// NewNoopLogger creates a no-op logger that safely discards all log messages
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// Pre-define common fields for reuse

// Str creates a string field for structured logging
func Str(key string, val string) zap.Field {
	return zap.String(key, val)
}

// Int creates an integer field for structured logging
func Int(key string, val int) zap.Field {
	return zap.Int(key, val)
}

// Bool creates a boolean field for structured logging
func Bool(key string, val bool) zap.Field {
	return zap.Bool(key, val)
}

// Err creates an error field for structured logging
// It formats the error message to replace newlines with spaces
// to prevent multi-line log output from joined errors
func Err(err error) zap.Field {
	if err == nil {
		return zap.String("error", "")
	}
	return zap.String("error", formatError(err))
}

// formatError formats error messages by replacing newlines and tabs with spaces
// This ensures that joined errors don't create multi-line log output
func formatError(err error) string {
	if err == nil {
		return ""
	}
	// Replace newlines and tabs with spaces to prevent multi-line log output
	return strings.ReplaceAll(strings.ReplaceAll(err.Error(), "\n", " | "), "\t", " | ")
}

// Package-level logging functions that use the default logger

// Debug logs a debug message using the default logger
func Debug(msg string, fields ...zap.Field) {
	defaultLogger.Debug(msg, fields...)
}

// Info logs an info message using the default logger
func Info(msg string, fields ...zap.Field) {
	defaultLogger.Info(msg, fields...)
}

// Warn logs a warning message using the default logger
func Warn(msg string, fields ...zap.Field) {
	defaultLogger.Warn(msg, fields...)
}

// Error logs an error message using the default logger
func Error(msg string, fields ...zap.Field) {
	defaultLogger.Error(msg, fields...)
}

// Fatal logs a fatal message using the default logger
func Fatal(msg string, fields ...zap.Field) {
	defaultLogger.Fatal(msg, fields...)
}

// Logger implementation methods

// Debug logs a debug message
func (l *SugarLogger) Debug(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Debug(msg, fields...)
	}
}

// Info logs an info message
func (l *SugarLogger) Info(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Info(msg, fields...)
	}
}

// Warn logs a warning message
func (l *SugarLogger) Warn(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Warn(msg, fields...)
	}
}

// Error logs an error message
func (l *SugarLogger) Error(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func (l *SugarLogger) Fatal(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Fatal(msg, fields...)
	}
}

// Zap returns the underlying zap logger if needed
func (l *SugarLogger) Zap() *zap.Logger {
	return l.Log
}

// Log level constants
var (
	DebugLevel = zapcore.DebugLevel
	InfoLevel  = zapcore.InfoLevel
	WarnLevel  = zapcore.WarnLevel
	ErrorLevel = zapcore.ErrorLevel
	FatalLevel = zapcore.FatalLevel
)

// DisabledLevel returns a level that disables all logging
func DisabledLevel() zapcore.Level {
	return zapcore.Level(127)
}
