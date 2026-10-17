// Package bootstrap 负责按顺序装配配置、日志、数据库、模板、路由和 HTTP 服务。
package bootstrap

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"cms/internal/config"
)

// NewLogger 创建控制台彩色 + 文件 JSON 双输出的 slog.Logger。
func NewLogger(cfg *config.Config) (*slog.Logger, io.Closer, error) {
	level := parseLevel(cfg.Log.Level)
	console := &consoleHandler{level: level, out: os.Stdout, mu: &sync.Mutex{}}

	if cfg.Log.Dir == "" {
		return slog.New(console), nil, nil
	}
	if err := os.MkdirAll(cfg.Log.Dir, 0o750); err != nil {
		return nil, nil, fmt.Errorf("create log dir: %w", err)
	}
	file, err := os.OpenFile(filepath.Join(cfg.Log.Dir, cfg.Log.File), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o640)
	if err != nil {
		return nil, nil, fmt.Errorf("open log file: %w", err)
	}
	fileHandler := slog.NewJSONHandler(file, &slog.HandlerOptions{Level: level})
	handler := &multiHandler{handlers: []slog.Handler{console, fileHandler}}
	return slog.New(handler), file, nil
}

func parseLevel(s string) slog.Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return slog.LevelDebug
	case "warn", "warning":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

type multiHandler struct {
	handlers []slog.Handler
}

func (m *multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m *multiHandler) Handle(ctx context.Context, r slog.Record) error {
	for _, h := range m.handlers {
		if h.Enabled(ctx, r.Level) {
			if err := h.Handle(ctx, r.Clone()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := make([]slog.Handler, 0, len(m.handlers))
	for _, h := range m.handlers {
		next = append(next, h.WithAttrs(attrs))
	}
	return &multiHandler{handlers: next}
}

func (m *multiHandler) WithGroup(name string) slog.Handler {
	next := make([]slog.Handler, 0, len(m.handlers))
	for _, h := range m.handlers {
		next = append(next, h.WithGroup(name))
	}
	return &multiHandler{handlers: next}
}

// consoleHandler 是带 ANSI 颜色的轻量控制台 handler。
type consoleHandler struct {
	level slog.Level
	out   io.Writer
	attrs []slog.Attr
	group string
	mu    *sync.Mutex
}

func (h *consoleHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *consoleHandler) Handle(_ context.Context, r slog.Record) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	color := "\x1b[36m"
	switch {
	case r.Level >= slog.LevelError:
		color = "\x1b[31m"
	case r.Level >= slog.LevelWarn:
		color = "\x1b[33m"
	case r.Level >= slog.LevelInfo:
		color = "\x1b[32m"
	}
	var b strings.Builder
	b.WriteString(r.Time.Format("2006-01-02 15:04:05.000"))
	b.WriteByte(' ')
	b.WriteString(color)
	b.WriteString(strings.ToUpper(r.Level.String()))
	b.WriteString("\x1b[0m ")
	b.WriteString(r.Message)
	writeAttrs := func(attrs []slog.Attr) {
		for _, a := range attrs {
			if h.group != "" {
				b.WriteString(" ")
				b.WriteString(h.group)
				b.WriteByte('.')
				b.WriteString(a.Key)
			} else {
				b.WriteString(" ")
				b.WriteString(a.Key)
			}
			b.WriteByte('=')
			b.WriteString(a.Value.String())
		}
	}
	writeAttrs(h.attrs)
	r.Attrs(func(a slog.Attr) bool {
		writeAttrs([]slog.Attr{a})
		return true
	})
	b.WriteByte('\n')
	_, err := io.WriteString(h.out, b.String())
	return err
}

func (h *consoleHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	next := *h
	next.attrs = append(append([]slog.Attr{}, h.attrs...), attrs...)
	return &next
}

func (h *consoleHandler) WithGroup(name string) slog.Handler {
	next := *h
	if next.group != "" {
		next.group += "." + name
	} else {
		next.group = name
	}
	return &next
}
