package logger

import (
	"go.uber.org/zap"
	zapCore "go.uber.org/zap/zapcore"
)

var Logger *zap.SugaredLogger

func Init() {
	cfg := zap.NewDevelopmentConfig()
	cfg.Encoding = "console"
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.LevelKey = "level"
	cfg.EncoderConfig.CallerKey = "caller"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.EncodeLevel = zapCore.CapitalColorLevelEncoder // 색상
	cfg.EncoderConfig.EncodeTime = zapCore.ISO8601TimeEncoder

	base, _ := cfg.Build()
	Logger = base.Sugar()
}
