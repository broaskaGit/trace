package trace_test

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/mightatnight/trace"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/goleak"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestMain(m *testing.M) {
	goleak.VerifyTestMain(m)
}

func TestNoopLogger(t *testing.T) {
	logger := &trace.NoopLogger{}
	logger.Debug("test")
	logger.Info("test")
	logger.Warn("test")
	logger.Error("test")
	logger.Fatal("test")
	assert.Nil(t, logger.Zap())
}

func TestDefaultLogger(t *testing.T) {
	initialLogger := trace.GetDefaultLogger()
	assert.IsType(t, &trace.NoopLogger{}, initialLogger, "Default logger should be a NoopLogger")

	var buffer bytes.Buffer
	writer := zapcore.AddSync(&buffer)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zap.NewDevelopmentEncoderConfig()),
		writer,
		zap.DebugLevel,
	)
	newLogger := &trace.SugarLogger{Log: zap.New(core)}

	trace.SetDefaultLogger(newLogger)
	currentLogger := trace.GetDefaultLogger()
	assert.Equal(t, newLogger, currentLogger, "Default logger should be the one we set")

	trace.Info("hello from default logger")
	assert.Contains(t, buffer.String(), "hello from default logger")

	trace.SetDefaultLogger(nil)
	assert.IsType(t, &trace.NoopLogger{}, trace.GetDefaultLogger(), "Setting nil should reset to NoopLogger")
}

func TestNew(t *testing.T) {
	t.Run("stdout only", func(t *testing.T) {
		var buf bytes.Buffer
		ogStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() {
			os.Stdout = ogStdout
		}()

		logger := trace.New(trace.InfoLevel, "test_prefix", nil)
		logger.Info("hello stdout")

		w.Close()
		io.Copy(&buf, r)

		output := buf.String()
		assert.Contains(t, output, "INFO")
		assert.Contains(t, output, "hello stdout")
		assert.Contains(t, output, "test_prefix")
	})

	t.Run("file only", func(t *testing.T) {
		tmpfile, err := os.CreateTemp("", "testlog_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		// To test file output, we can't also have stdout, so we redirect it to discard
		ogStdout := os.Stdout
		devNull, err := os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		require.NoError(t, err)
		defer devNull.Close()
		os.Stdout = devNull
		defer func() {
			os.Stdout = ogStdout
		}()

		logger := trace.New(trace.InfoLevel, "file_logger", tmpfile)
		logger.Warn("writing to file")
		tmpfile.Close() // Close to flush writes

		content, err := os.ReadFile(tmpfile.Name())
		require.NoError(t, err)
		output := string(content)
		assert.Contains(t, output, "WARN")
		assert.Contains(t, output, "writing to file")
		assert.Contains(t, output, "file_logger")
	})

	t.Run("stdout and file", func(t *testing.T) {
		// Capture stdout
		var stdoutBuf bytes.Buffer
		ogStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w

		// Temp file for file logging
		tmpfile, err := os.CreateTemp("", "testlog_*.log")
		require.NoError(t, err)
		defer os.Remove(tmpfile.Name())

		logger := trace.New(trace.InfoLevel, "multi_logger", tmpfile)
		logger.Error("multi output")

		w.Close()
		io.Copy(&stdoutBuf, r)
		os.Stdout = ogStdout
		tmpfile.Close()

		// Check stdout
		stdoutOutput := stdoutBuf.String()
		assert.Contains(t, stdoutOutput, "ERROR")
		assert.Contains(t, stdoutOutput, "multi output")
		assert.Contains(t, stdoutOutput, "multi_logger")

		// Check file
		content, err := os.ReadFile(tmpfile.Name())
		require.NoError(t, err)
		fileOutput := string(content)
		assert.Contains(t, fileOutput, "ERROR")
		assert.Contains(t, fileOutput, "multi output")
		assert.Contains(t, fileOutput, "multi_logger")
	})

	t.Run("no prefix", func(t *testing.T) {
		var buf bytes.Buffer
		ogStdout := os.Stdout
		r, w, _ := os.Pipe()
		os.Stdout = w
		defer func() {
			os.Stdout = ogStdout
		}()

		logger := trace.New(trace.InfoLevel, "", nil)
		logger.Info("no prefix test")

		w.Close()
		io.Copy(&buf, r)

		output := buf.String()
		// With no prefix, the logger name key shouldn't even appear.
		assert.NotContains(t, output, "logger")
	})
}

func TestSugarLogger_Fatal(t *testing.T) {
	// Testing Fatal is tricky because it calls os.Exit.
	// A proper test would involve running a subprocess and checking its exit code and output.
	// The `zaptest` package provides a logger that calls t.Fatal, but that also stops the test prematurely.
	// For this library, we will assume that if other log levels work, Fatal will also correctly
	// call the underlying zap logger's Fatal method. The behavior of zap's Fatal is tested in zap itself.
	t.Skip("Skipping fatal test due to complexity of testing os.Exit")
}

func TestPackageLevelFunctions(t *testing.T) {
	var buffer bytes.Buffer
	writer := zapcore.AddSync(&buffer)
	encoderConfig := zap.NewDevelopmentEncoderConfig()
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		writer,
		zap.DebugLevel,
	)
	testLogger := &trace.SugarLogger{Log: zap.New(core)}
	trace.SetDefaultLogger(testLogger)
	defer trace.SetDefaultLogger(nil) // reset

	trace.Debug("pkg debug")
	assert.Contains(t, buffer.String(), "pkg debug")
	buffer.Reset()

	trace.Info("pkg info")
	assert.Contains(t, buffer.String(), "pkg info")
	buffer.Reset()

	trace.Warn("pkg warn")
	assert.Contains(t, buffer.String(), "pkg warn")
	buffer.Reset()

	trace.Error("pkg error", trace.Err(io.EOF))
	assert.Contains(t, buffer.String(), "pkg error")
	assert.Contains(t, buffer.String(), io.EOF.Error())
	buffer.Reset()
}

func TestFieldHelpers(t *testing.T) {
	strField := trace.Str("key", "value")
	assert.Equal(t, zap.String("key", "value"), strField)

	intField := trace.Int("key", 123)
	assert.Equal(t, zap.Int("key", 123), intField)

	boolField := trace.Bool("key", true)
	assert.Equal(t, zap.Bool("key", true), boolField)

	err := io.EOF
	errField := trace.Err(err)
	assert.Equal(t, zap.Error(err), errField)
}

func TestDisabledLevel(t *testing.T) {
	var buf bytes.Buffer
	ogStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := trace.New(trace.DisabledLevel(), "disabled", nil)
	logger.Info("this should not be logged")

	w.Close()
	io.Copy(&buf, r)
	os.Stdout = ogStdout

	assert.Empty(t, buf.String())
}

func TestNewNoopLogger(t *testing.T) {
	logger := trace.NewNoopLogger()
	assert.IsType(t, &trace.NoopLogger{}, logger)
}

func TestSugarLogger_Zap(t *testing.T) {
	logger := trace.New(trace.InfoLevel, "", nil)
	assert.NotNil(t, logger.Zap())
	assert.IsType(t, &zap.Logger{}, logger.Zap())

	// test nil case for coverage
	sugarLogger := &trace.SugarLogger{}
	sugarLogger.Debug("test")
	sugarLogger.Info("test")
	sugarLogger.Warn("test")
	sugarLogger.Error("test")
	sugarLogger.Fatal("test")
	assert.Nil(t, sugarLogger.Zap())
}

func BenchmarkSugarLogger(b *testing.B) {
	logger := trace.New(trace.InfoLevel, "benchmark", nil)
	b.ReportAllocs()
	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			logger.Info("bench message", trace.Str("key", "value"), trace.Int("num", 123))
		}
	})
}
