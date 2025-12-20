package logger

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/contrib/bridges/otelslog"
	"go.opentelemetry.io/otel/trace"
)

// Config holds logger configuration
type Config struct {
	Level       string // debug, info, warn, error
	Format      string // json, text
	ServiceName string
}

// InitLogger initializes the global logger with OTel support
func InitLogger(cfg Config) {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.TimeKey {
				return slog.String(slog.TimeKey, a.Value.Time().Format(time.RFC3339))
			}
			return a
		},
	}

	// 1. Stdout Handler (with Trace Context support)
	var baseHandler slog.Handler
	if cfg.Format == "json" {
		baseHandler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		baseHandler = slog.NewTextHandler(os.Stdout, opts)
	}
	stdoutHandler := &TraceContextHandler{Handler: baseHandler}

	// 2. OTel Handler
	// otelslog handler extracts trace context automatically
	otelHandler := otelslog.NewHandler(cfg.ServiceName)

	// 3. Fanout Handler
	fanout := NewFanoutHandler(stdoutHandler, otelHandler)

	// Set as global default
	logger := slog.New(fanout)
	slog.SetDefault(logger)
}

// TraceContextHandler adds trace/span IDs to the record from context
type TraceContextHandler struct {
	slog.Handler
}

func (h *TraceContextHandler) Handle(ctx context.Context, r slog.Record) error {
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		r.AddAttrs(
			slog.String("trace_id", span.SpanContext().TraceID().String()),
			slog.String("span_id", span.SpanContext().SpanID().String()),
		)
	}
	return h.Handler.Handle(ctx, r)
}

func (h *TraceContextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &TraceContextHandler{Handler: h.Handler.WithAttrs(attrs)}
}

func (h *TraceContextHandler) WithGroup(name string) slog.Handler {
	return &TraceContextHandler{Handler: h.Handler.WithGroup(name)}
}

// FanoutHandler sends log records to multiple handlers
type FanoutHandler struct {
	handlers []slog.Handler
}

func NewFanoutHandler(handlers ...slog.Handler) slog.Handler {
	return &FanoutHandler{handlers: handlers}
}

func (h *FanoutHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (h *FanoutHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, handler := range h.handlers {
		if handler.Enabled(ctx, r.Level) {
			// Ignore errors from individual handlers to ensure best effort
			_ = handler.Handle(ctx, r)
		}
	}
	return nil
}

func (h *FanoutHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithAttrs(attrs)
	}
	return NewFanoutHandler(handlers...)
}

func (h *FanoutHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(h.handlers))
	for i, handler := range h.handlers {
		handlers[i] = handler.WithGroup(name)
	}
	return NewFanoutHandler(handlers...)
}
