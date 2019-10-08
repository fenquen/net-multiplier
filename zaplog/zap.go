package zaplog

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net-multiplier/config"
)

var LOGGER *zap.Logger

func InitZapLog() {
	cfg := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.DebugLevel),
		Development: true,
		Encoding:    "console",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "trace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.CapitalLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout", "./" + config.APP_NAME + ".log"},
		ErrorOutputPaths: []string{"stderr", "./" + config.APP_NAME + ".log"},
		/*InitialFields: map[string]interface{}{
			"app": "test",
		},*/
	}

	LOGGER, _ = cfg.Build()
	if nil == LOGGER {
		panic("zaplog init fail:")
	}
}

func init() {
	InitZapLog()
}

func Info(msg string, fields ...zap.Field) {
	LOGGER.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	LOGGER.Error(msg, fields...)
}
