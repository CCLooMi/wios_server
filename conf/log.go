package conf

import (
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// zapWriter is a custom io.Writer that writes to a zap logger
type zapWriter struct {
	logger *zap.Logger
	lv     zapcore.Level
}

// Write implements the io.Writer interface for zapWriter
func (w zapWriter) Write(p []byte) (n int, err error) {
	pL := len(p)
	if w.lv == zap.DebugLevel {
		w.logger.Debug(string(p))
		return pL, nil
	}
	if w.lv == zap.InfoLevel {
		w.logger.Info(string(p))
		return pL, nil
	}
	if w.lv == zap.WarnLevel {
		w.logger.Warn(string(p))
		return pL, nil
	}
	if w.lv == zap.ErrorLevel {
		w.logger.Error(string(p))
		return pL, nil
	}
	return pL, nil
}
func setLog(config *Config) *zap.Logger {
	zapCfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: false,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "ts",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05"),
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}
	logger, err := zapCfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flushes buffer, if any
	// Replace the global logger
	// Replace zap's global logger
	zap.ReplaceGlobals(logger)
	// Redirect stdlib log output to our logger
	zap.RedirectStdLog(logger)
	// Set Gin to use zap's logger
	gin.DefaultWriter = zapWriter{logger: logger, lv: zap.DebugLevel}
	gin.DefaultErrorWriter = zapWriter{logger: logger, lv: zap.ErrorLevel}
	// Set log level from configuration
	logLevel, err := zapcore.ParseLevel(config.LogLevel)
	if err != nil {
		logLevel = zapcore.DebugLevel
	}
	zapCfg.Level.SetLevel(logLevel)
	return logger
}
