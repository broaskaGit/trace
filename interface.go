package trace

import "go.uber.org/zap"

// Logger defines the logging methods
type Logger interface {
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
	// With returns a child logger with additional structured fields included in every log.
	With(fields ...zap.Field) Logger
	// Named returns a child logger with a name scope (logger name prefix).
	Named(name string) Logger
	// Zap returns the underlying zap.Logger.
	Zap() *zap.Logger
}

// NoopLogger implements LoggerInterface with no-op operations
type NoopLogger struct{}

// NewNoopLogger creates a no-op logger that safely discards all log messages
func NewNoopLogger() Logger {
	return &NoopLogger{}
}

// NoopLogger implementation methods

func (n *NoopLogger) Debug(msg string, fields ...zap.Field) {}
func (n *NoopLogger) Info(msg string, fields ...zap.Field)  {}
func (n *NoopLogger) Warn(msg string, fields ...zap.Field)  {}
func (n *NoopLogger) Error(msg string, fields ...zap.Field) {}
func (n *NoopLogger) Fatal(msg string, fields ...zap.Field) {}
func (n *NoopLogger) With(fields ...zap.Field) Logger       { return n }
func (n *NoopLogger) Named(name string) Logger              { return n }
func (n *NoopLogger) Zap() *zap.Logger                      { return zap.NewNop() }
