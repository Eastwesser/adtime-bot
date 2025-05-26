package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ZAP LOGGER

func New() (*zap.Logger, error) {
	config := zap.NewProductionConfig()
	config.EncoderConfig.TimeKey = "timestamp"
	config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	return config.Build()
}
