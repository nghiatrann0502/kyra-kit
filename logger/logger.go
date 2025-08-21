package logger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Extractor struct {
	name      string
	extractor func(context.Context) any
}

type ZapHandler struct {
	logger *zap.Logger
	attrs  []slog.Attr
}

func NewZapHandler(cfg zap.Config) (*ZapHandler, error) {
	logger, err := cfg.Build()
	if err != nil {
		return nil, err
	}

	return &ZapHandler{
		logger: logger,
		attrs:  []slog.Attr{},
	}, nil
}

func NewZapHandlerWithLogger(logger *zap.Logger) *ZapHandler {
	return &ZapHandler{
		logger: logger,
		attrs:  []slog.Attr{},
	}
}

func (h *ZapHandler) Enabled(_ context.Context, level slog.Level) bool {
	zapLevel := slogLevelToZapLevel(level)
	return h.logger.Core().Enabled(zapLevel)
}

func (h *ZapHandler) Handle(ctx context.Context, record slog.Record) error {
	if !h.Enabled(ctx, record.Level) {
		return nil
	}

	zapLevel := slogLevelToZapLevel(record.Level)

	fields := make([]zap.Field, 0, record.NumAttrs()+len(h.attrs))

	for _, attr := range h.attrs {
		fields = append(fields, slogAttrToZapField(attr))
	}

	record.Attrs(func(attr slog.Attr) bool {
		fields = append(fields, slogAttrToZapField(attr))
		return true
	})

	if ctx != nil {
		fields = append(fields, extractContextValues(ctx)...)
	}

	ce := h.logger.Check(zapLevel, record.Message)
	if ce != nil {
		ce.Write(fields...)
	}

	return nil
}

func (h *ZapHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandler := *h
	newHandler.attrs = append(h.attrs, attrs...)
	return &newHandler
}

func (h *ZapHandler) WithGroup(name string) slog.Handler {
	newHandler := *h
	if name == "" {
		return &newHandler
	}

	newAttrs := make([]slog.Attr, 0, len(h.attrs))
	for _, attr := range h.attrs {
		newAttrs = append(newAttrs, slog.Attr{
			Key:   name + "." + attr.Key,
			Value: attr.Value,
		})
	}
	newHandler.attrs = newAttrs
	return &newHandler
}

func slogLevelToZapLevel(level slog.Level) zapcore.Level {
	switch {
	case level >= slog.LevelError:
		return zapcore.ErrorLevel
	case level >= slog.LevelWarn:
		return zapcore.WarnLevel
	case level >= slog.LevelInfo:
		return zapcore.InfoLevel
	default:
		return zapcore.DebugLevel
	}
}

func slogAttrToZapField(attr slog.Attr) zap.Field {
	key := attr.Key
	value := attr.Value
	kind := attr.Value.Kind()

	switch kind {
	case slog.KindBool:
		return zap.Bool(key, value.Bool())
	case slog.KindDuration:
		return zap.Duration(key, value.Duration())
	case slog.KindFloat64:
		return zap.Float64(key, value.Float64())
	case slog.KindInt64:
		return zap.Int64(key, value.Int64())
	case slog.KindString:
		return zap.String(key, value.String())
	case slog.KindTime:
		return zap.Time(key, value.Time())
	case slog.KindUint64:
		return zap.Uint64(key, value.Uint64())
	case slog.KindGroup:
		// For groups, we need to flatten the attributes
		attrs := value.Group()
		fields := make([]zap.Field, 0, len(attrs))
		for _, attr := range attrs {
			groupKey := key
			if groupKey != "" {
				groupKey += "."
			}
			groupKey += attr.Key

			fields = append(fields, slogAttrToZapField(slog.Attr{
				Key:   groupKey,
				Value: attr.Value,
			}))
		}
		if len(fields) == 1 {
			return fields[0]
		}
		return zap.Object(key, zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
			for _, field := range fields {
				field.AddTo(enc)
			}
			return nil
		}))
	default:
		return zap.String(key, value.String())
	}
}

func NewSlogLogger(zapLogger *zap.Logger) *slog.Logger {
	handler := NewZapHandlerWithLogger(zapLogger)
	return slog.New(handler)
}

func DefaultSlogLogger() (*slog.Logger, error) {
	cfg := zap.NewProductionConfig()
	handler, err := NewZapHandler(cfg)
	if err != nil {
		return nil, err
	}
	return slog.New(handler), nil
}

func extractContextValues(ctx context.Context) []zap.Field {
	if ctx == nil {
		return nil
	}

	var fields []zap.Field

	extractors := []Extractor{
		// {"request_id", func(ctx context.Context) any {
		// 	return contextutils.RequestIDFromContext(ctx)
		// }},
		// {"ip_addr", func(ctx context.Context) any {
		// 	if val, ok := ctx.Value(utils.CtxIPKey).(string); ok {
		// 		return val
		// 	}
		// 	return nil
		// }},
	}

	for _, ext := range extractors {
		if val := ext.extractor(ctx); val != nil {
			switch v := val.(type) {
			case string:
				fields = append(fields, zap.String(ext.name, v))
			case int:
				fields = append(fields, zap.Int(ext.name, v))
			case int64:
				fields = append(fields, zap.Int64(ext.name, v))
			case float64:
				fields = append(fields, zap.Float64(ext.name, v))
			case bool:
				fields = append(fields, zap.Bool(ext.name, v))
			case time.Time:
				fields = append(fields, zap.Time(ext.name, v))
			default:
				fields = append(fields, zap.String(ext.name, fmt.Sprintf("%v", v)))
			}
		}
	}

	return fields
}

func InitProvider(isProduction bool) *zap.Logger {
	encoderConfig := zapcore.EncoderConfig{
		TimeKey:     "ts",
		LevelKey:    "level",
		NameKey:     "logger",
		CallerKey:   "caller",
		FunctionKey: zapcore.OmitKey,
		MessageKey:  "msg",
		// StacktraceKey:  "stack_trace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.LowercaseLevelEncoder,
		EncodeTime:     zapcore.ISO8601TimeEncoder,
		EncodeDuration: zapcore.StringDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	zConfig := zap.Config{
		Level:       zap.NewAtomicLevelAt(zap.InfoLevel),
		Development: !isProduction,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    encoderConfig,
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if !isProduction {
		zConfig.Encoding = "console"
	}

	zapLogger, err := zConfig.Build()
	if err != nil {
		panic(err)
	}
	defer zapLogger.Sync()

	return zapLogger
}
