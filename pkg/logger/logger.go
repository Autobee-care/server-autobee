// Package logger provides a production-ready Zap logger factory.
package logger

import (
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// New creates and returns a configured Zap logger.
// In development it uses a console-friendly encoder; in production it uses JSON.
func New(level, env string) (*zap.Logger, error) {
	var zapLevel zapcore.Level
	if err := zapLevel.UnmarshalText([]byte(level)); err != nil {
		zapLevel = zapcore.InfoLevel
	}

	var cfg zap.Config
	if env == "production" {
		cfg = zap.NewProductionConfig()
	} else {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	cfg.Level = zap.NewAtomicLevelAt(zapLevel)

	log, err := cfg.Build(zap.AddCallerSkip(0))
	if err != nil {
		return nil, err
	}
	return log, nil
}
