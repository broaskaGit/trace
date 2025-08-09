package trace

import "go.uber.org/zap"

// Logger defines the logging methods
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	Zap() *zap.Logger
}

// NoopLogger implements LoggerInterface with no-op operations
type NoopLogger struct{}

// NoopLogger implementation methods

// Debug is a no-op
func (n *NoopLogger) Debug(msg string, fields ...zap.Field) {}

// Info is a no-op
func (n *NoopLogger) Info(msg string, fields ...zap.Field) {}

// Warn is a no-op
func (n *NoopLogger) Warn(msg string, fields ...zap.Field) {}

// Error is a no-op
func (n *NoopLogger) Error(msg string, fields ...zap.Field) {}

// Fatal is a no-op
func (n *NoopLogger) Fatal(msg string, fields ...zap.Field) {}

// Zap returns nil for the underlying zap logger
func (n *NoopLogger) Zap() *zap.Logger {
	return nil
}

// Default global logger (no-op by default)
var defaultLogger Logger = &NoopLogger{}

// SetDefaultLogger sets the global default logger
func SetDefaultLogger(logger Logger) {
	if logger != nil {
		defaultLogger = logger
	} else {
		defaultLogger = &NoopLogger{}
	}
}

// GetDefaultLogger returns the current global default logger
func GetDefaultLogger() Logger {
	return defaultLogger
}
