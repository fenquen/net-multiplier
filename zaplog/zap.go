package zaplog

import (
	"fmt"
	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net-multiplier/config"
)

var LOGGER *zap.Logger
var logLevel zap.AtomicLevel

func InitZapLogger() {

	cfg := zap.Config{
		Level:       logLevel,
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

/**
 * 获取日志
 * filePath 日志文件路径
 * level 日志级别
 * maxSize 每个日志文件保存的最大尺寸 单位：M
 * maxBackups 日志文件最多保存多少个备份
 * maxAge 文件最多保存多少天
 * compress 是否压缩
 * serviceName 服务名
 */
func InitZapLoggerLumber() {
	hook := &lumberjack.Logger{
		Filename:   fmt.Sprintf("./" + config.APP_NAME + ".log"),
		MaxSize:    10,
		MaxBackups: 10000,
		MaxAge:     100000,
		Compress:   true,
	}
	//defer hook.Close()

	level := logLevel
	w := zapcore.AddSync(hook)
	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(zapcore.EncoderConfig{
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
		}),    //编码器配置
		w,     //打印到控制台和文件
		level, //日志等级
	)

	// 后边还能添加 zap.AddCallerSkip(1)
	LOGGER = zap.New(core, zap.AddCaller())
}
func init() {
	switch *config.LogLevel {
	case "info":
		logLevel = zap.NewAtomicLevelAt(zap.InfoLevel)
	case "debug":
		logLevel = zap.NewAtomicLevelAt(zap.DebugLevel)
	}
	InitZapLogger()
}

func Info(msg string, fields ...zap.Field) {
	LOGGER.Info(msg, fields...)
}

func Error(msg string, fields ...zap.Field) {
	LOGGER.Error(msg, fields...)
}
