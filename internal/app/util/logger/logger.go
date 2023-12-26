package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var Log = zap.NewNop()

// Initialize func initializing new Log instance.
func Initialize(level string) error {
	lvl, err := zap.ParseAtomicLevel(level)
	if err != nil {
		return fmt.Errorf("initialize parse level %w", err)
	}

	cfg := zap.NewProductionConfig()
	cfg.Level = lvl

	cfg.EncoderConfig.LineEnding = zapcore.DefaultLineEnding
	cfg.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncoderConfig.EncodeDuration = zapcore.StringDurationEncoder

	zl, err := cfg.Build()
	if err != nil {
		return fmt.Errorf("initialize cfg build %w", err)
	}
	Log = zl
	return nil
}
