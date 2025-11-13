package log

import (
	"os"
	"path/filepath"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	// ContextDir is the directory where cache files and logs are stored
	ContextDir = "/tmp/bw-mcp"
)

var (
	// log is the global sugared logger instance
	log *zap.SugaredLogger
)

// Init initializes the global logger
// It creates a production logger that writes JSON logs to /tmp/bw-mcp/log.txt
// The log file is truncated on each startup
func Init() error {
	// Ensure the context directory exists
	if err := os.MkdirAll(ContextDir, 0755); err != nil {
		return err
	}

	// Create log file path
	logPath := filepath.Join(ContextDir, "log.txt")

	// Open log file with truncation (overwrite on startup)
	logFile, err := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}

	// Create encoder config for production
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	// Create file writer core
	fileEncoder := zapcore.NewJSONEncoder(encoderConfig)
	fileCore := zapcore.NewCore(
		fileEncoder,
		zapcore.AddSync(logFile),
		zapcore.InfoLevel, // Info level and above
	)

	// Create logger with file output only
	baseLogger := zap.New(fileCore, zap.AddCaller(), zap.AddStacktrace(zapcore.ErrorLevel))
	log = baseLogger.Sugar()

	return nil
}

// Logger returns the global sugared logger instance
func Logger() *zap.SugaredLogger {
	return log
}

// Sync flushes any buffered log entries
func Sync() {
	if log != nil {
		_ = log.Sync()
	}
}
