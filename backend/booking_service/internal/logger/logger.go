package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"sync"
)

const (
	timeFormat   = "[15:04:05.000]"
	reset        = "\033[0m"
	cyan         = 36
	lightGray    = 37
	darkGray     = 90
	lightRed     = 91
	lightGreen   = 92
	lightYellow  = 93
	lightBlue    = 94
	lightMagenta = 95
	lightCyan    = 96
	white        = 97
)

const (
	FormatText = "text"
	FormatJSON = "json"
)

func colorizer(colorCode int, v string) string {
	return fmt.Sprintf("\033[%sm%s%s", strconv.Itoa(colorCode), v, reset)
}

type Handler struct {
	h                slog.Handler
	r                func([]string, slog.Attr) slog.Attr
	b                *bytes.Buffer
	m                *sync.Mutex
	writer           io.Writer
	colorize         bool
	outputEmptyAttrs bool
	outputFormat     string
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.h.Enabled(ctx, level)
}

func (h *Handler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &Handler{
		h:                h.h.WithAttrs(attrs),
		b:                h.b,
		r:                h.r,
		m:                h.m,
		writer:           h.writer,
		colorize:         h.colorize,
		outputEmptyAttrs: h.outputEmptyAttrs,
		outputFormat:     h.outputFormat,
	}
}

func (h *Handler) WithGroup(name string) slog.Handler {
	return &Handler{
		h:                h.h.WithGroup(name),
		b:                h.b,
		r:                h.r,
		m:                h.m,
		writer:           h.writer,
		colorize:         h.colorize,
		outputEmptyAttrs: h.outputEmptyAttrs,
		outputFormat:     h.outputFormat,
	}
}

func (h *Handler) computeAttrs(ctx context.Context, r slog.Record) (map[string]any, error) {
	h.m.Lock()
	defer func() {
		h.b.Reset()
		h.m.Unlock()
	}()
	if err := h.h.Handle(ctx, r); err != nil {
		return nil, fmt.Errorf("error when calling inner handler's Handle: %w", err)
	}

	var attrs map[string]any
	err := json.Unmarshal(h.b.Bytes(), &attrs)
	if err != nil {
		return nil, fmt.Errorf("error when unmarshaling inner handler's Handle result: %w", err)
	}
	return attrs, nil
}

func (h *Handler) Handle(ctx context.Context, r slog.Record) error {
	colorize := func(code int, value string) string {
		return value
	}
	if h.colorize {
		colorize = colorizer
	}

	var level string
	levelAttr := slog.Attr{
		Key:   slog.LevelKey,
		Value: slog.AnyValue(r.Level),
	}
	if h.r != nil {
		levelAttr = h.r([]string{}, levelAttr)
	}

	if !levelAttr.Equal(slog.Attr{}) {
		level = levelAttr.Value.String() + ":"

		if r.Level <= slog.LevelDebug {
			level = colorize(lightGreen, level)
		} else if r.Level <= slog.LevelInfo {
			level = colorize(cyan, level)
		} else if r.Level < slog.LevelWarn {
			level = colorize(lightBlue, level)
		} else if r.Level < slog.LevelError {
			level = colorize(lightYellow, level)
		} else if r.Level <= slog.LevelError+1 {
			level = colorize(lightRed, level)
		} else if r.Level > slog.LevelError+1 {
			level = colorize(lightMagenta, level)
		}
	}

	var timestamp string
	timeAttr := slog.Attr{
		Key:   slog.TimeKey,
		Value: slog.StringValue(r.Time.Format(timeFormat)),
	}
	if h.r != nil {
		timeAttr = h.r([]string{}, timeAttr)
	}
	if !timeAttr.Equal(slog.Attr{}) {
		timestamp = colorize(lightGray, timeAttr.Value.String())
	}

	var msg string
	msgAttr := slog.Attr{
		Key:   slog.MessageKey,
		Value: slog.StringValue(r.Message),
	}
	if h.r != nil {
		msgAttr = h.r([]string{}, msgAttr)
	}
	if !msgAttr.Equal(slog.Attr{}) {
		msg = colorize(white, msgAttr.Value.String())
	}

	switch h.outputFormat {
	case FormatJSON:
		return h.handleJSON(ctx, r, timestamp, level, msg)
	default:
		return h.handleText(r, timestamp, level, msg)
	}
}

func (h *Handler) handleJSON(ctx context.Context, r slog.Record, timestamp, level, msg string) error {
	attrs, err := h.computeAttrs(ctx, r)
	if err != nil {
		return err
	}

	var attrsAsBytes []byte
	if h.outputEmptyAttrs || len(attrs) > 0 {
		attrsAsBytes, err = json.MarshalIndent(attrs, "", "  ")
		if err != nil {
			return fmt.Errorf("error when marshaling attrs: %w", err)
		}
	}

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
		out.WriteString(" ")
	}
	if len(msg) > 0 {
		out.WriteString(msg)
		out.WriteString(" ")
	}
	if len(attrsAsBytes) > 0 {
		out.WriteString(colorizer(darkGray, string(attrsAsBytes)))
	}

	_, err = io.WriteString(h.writer, out.String()+"\n")
	return err
}

func (h *Handler) handleText(r slog.Record, timestamp, level, msg string) error {
	var attrsStr string
	if h.outputEmptyAttrs || r.NumAttrs() > 0 {
		var attrsBuilder strings.Builder
		r.Attrs(func(a slog.Attr) bool {
			if attrsBuilder.Len() > 0 {
				attrsBuilder.WriteString(" ")
			}
			attrsBuilder.WriteString(colorizer(darkGray, a.Key+"="+a.Value.String()))
			return true
		})
		attrsStr = attrsBuilder.String()
	}

	out := strings.Builder{}
	if len(timestamp) > 0 {
		out.WriteString(timestamp)
		out.WriteString(" ")
	}
	if len(level) > 0 {
		out.WriteString(level)
		out.WriteString(" ")
	}
	if len(msg) > 0 {
		out.WriteString(msg)
	}
	if len(attrsStr) > 0 {
		if len(msg) > 0 {
			out.WriteString(" ")
		}
		out.WriteString(attrsStr)
	}

	_, err := io.WriteString(h.writer, out.String()+"\n")
	return err
}

func suppressDefaults(next func([]string, slog.Attr) slog.Attr) func([]string, slog.Attr) slog.Attr {
	return func(groups []string, a slog.Attr) slog.Attr {
		if a.Key == slog.TimeKey ||
			a.Key == slog.LevelKey ||
			a.Key == slog.MessageKey {
			return slog.Attr{}
		}
		if next == nil {
			return a
		}
		return next(groups, a)
	}
}

type Option func(h *Handler)

func WithDestinationWriter(writer io.Writer) Option {
	return func(h *Handler) {
		h.writer = writer
	}
}

func WithColor() Option {
	return func(h *Handler) {
		h.colorize = true
	}
}

func WithOutputEmptyAttrs() Option {
	return func(h *Handler) {
		h.outputEmptyAttrs = true
	}
}

func WithOutputFormat(format string) Option {
	return func(h *Handler) {
		h.outputFormat = format
	}
}

func New(handlerOptions *slog.HandlerOptions, options ...Option) *Handler {
	if handlerOptions == nil {
		handlerOptions = &slog.HandlerOptions{}
	}

	buf := &bytes.Buffer{}
	handler := &Handler{
		b: buf,
		h: slog.NewJSONHandler(buf, &slog.HandlerOptions{
			Level:       handlerOptions.Level,
			AddSource:   handlerOptions.AddSource,
			ReplaceAttr: suppressDefaults(handlerOptions.ReplaceAttr),
		}),
		r:            handlerOptions.ReplaceAttr,
		m:            &sync.Mutex{},
		outputFormat: FormatText,
	}

	for _, opt := range options {
		opt(handler)
	}

	return handler
}

func NewHandler(opts *slog.HandlerOptions) *Handler {
	return New(opts,
		WithDestinationWriter(os.Stdout),
		WithColor(),
		WithOutputEmptyAttrs(),
	)
}

func SetupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case "local":
		log = slog.New(New(
			&slog.HandlerOptions{Level: slog.LevelDebug},
			WithOutputFormat(FormatJSON),
			WithDestinationWriter(os.Stdout),
			WithColor(),
		))
	case "production":
		log = slog.New(NewHandler(&slog.HandlerOptions{Level: slog.LevelDebug}))
	}
	return log
}
