package util

import (
	"context"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type ctxKey string

const CtxKeyTraceID ctxKey = "trace_id"

func NewLogger(service string) (*zap.Logger, error) {
	cfg := zap.NewProductionConfig()
	// keep human-readable time in dev via Env
	if os.Getenv("ENV") != "production" {
		cfg = zap.NewDevelopmentConfig()
		cfg.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}
	cfg.EncoderConfig.TimeKey = "ts"
	cfg.EncoderConfig.MessageKey = "msg"
	cfg.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.InitialFields = map[string]interface{}{
		"service": service,
	}
	return cfg.Build()
}

// WithTraceFromContext returns a logger with the trace_id field if present
func WithTraceFromContext(ctx context.Context, logger *zap.Logger) *zap.Logger {
	if ctx == nil {
		return logger
	}
	if v := ctx.Value(CtxKeyTraceID); v != nil {
		if sid, ok := v.(string); ok && sid != "" {
			return logger.With(zap.String("trace_id", sid))
		}
	}
	return logger
}
