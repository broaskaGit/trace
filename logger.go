package trace

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var _ Logger = &sugarLogger{}

// sugarLogger implements the LoggerInterface with a real zap logger
type sugarLogger struct {
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
		return &sugarLogger{
			Log: zap.New(core).Named(prefix),
		}
	}

	// Build the logger with minimal options for speed
	return &sugarLogger{
		Log: zap.New(core),
	}
}

func NewChildLogger(parent Logger, prefix string) Logger {
	if parent == nil || parent.Zap() == nil {
		return NewNoopLogger()
	}

	if prefix != "" {
		return &sugarLogger{
			Log: parent.Zap().Named(prefix),
		}
	}

	return &sugarLogger{
		Log: parent.Zap(),
	}
}

// Debug logs a debug message
func (l *sugarLogger) Debug(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Debug(msg, fields...)
	}
}

// Info logs an info message
func (l *sugarLogger) Info(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Info(msg, fields...)
	}
}

// Warn logs a warning message
func (l *sugarLogger) Warn(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Warn(msg, fields...)
	}
}

// Error logs an error message
func (l *sugarLogger) Error(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Error(msg, fields...)
	}
}

// Fatal logs a fatal message and exits
func (l *sugarLogger) Fatal(msg string, fields ...zap.Field) {
	if l.Log != nil {
		l.Log.Fatal(msg, fields...)
	}
}

// With returns a child logger with additional structured fields included in every log.
func (l *sugarLogger) With(fields ...zap.Field) Logger {
	if l == nil || l.Log == nil {
		return l
	}
	return &sugarLogger{Log: l.Log.With(fields...)}
}

// Named returns a child logger with a name scope (logger name prefix).
func (l *sugarLogger) Named(name string) Logger {
	if l == nil || l.Log == nil {
		return l
	}
	return &sugarLogger{Log: l.Log.Named(name)}
}

// Zap returns the underlying zap logger if needed
func (l *sugarLogger) Zap() *zap.Logger {
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

// Context scoping helpers

type loggerCtxKey struct{}

// LoggerToContext attaches the provided logger to the context.
func LoggerToContext(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, loggerCtxKey{}, l)
}

// LoggerFromContext retrieves a logger from context or returns a no-op logger if absent.
// Typical usage:
//   - component-scoped: base := root.Named("http").With(zap.String("component","http"))
//   - request-scoped:  reqLog := base.With(zap.String("request_id", rid))
//   - ctx = WithLogger(ctx, reqLog)
func LoggerFromContext(ctx context.Context) Logger {
	if v := ctx.Value(loggerCtxKey{}); v != nil {
		if l, ok := v.(Logger); ok && l != nil && l.Zap() != nil {
			return l
		}
	}
	return NewNoopLogger()
}
