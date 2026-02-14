package logger

import (
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log *zap.Logger

func Init(level string) {
	writeSyncer := zapcore.AddSync(os.Stdout)
	encoder := getEncoder()
	
	var l zapcore.Level
	switch level {
	case "debug":
		l = zapcore.DebugLevel
	case "info":
		l = zapcore.InfoLevel
	case "warn":
		l = zapcore.WarnLevel
	case "error":
		l = zapcore.ErrorLevel
	default:
		l = zapcore.InfoLevel
	}

	core := zapcore.NewCore(encoder, writeSyncer, l)
	Log = zap.New(core, zap.AddCaller())
}

func getEncoder() zapcore.Encoder {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encoderConfig)
}
