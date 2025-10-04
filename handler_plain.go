package pho3nix_logger

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PlainTextHandler 是一个自定义的 Text Handler，输出不带 Key 的简洁格式
type PlainTextHandler struct {
	opts       slog.HandlerOptions
	mu         *sync.Mutex
	w          io.Writer
	prefix     string
	attrs      []slog.Attr
	timeFormat string
	levelMap   map[slog.Level]string
}

// NewPlainTextHandler 创建 PlainTextHandler
func NewPlainTextHandler(w io.Writer, opts *slog.HandlerOptions) *PlainTextHandler {
	h := &PlainTextHandler{w: w, mu: &sync.Mutex{}}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.Level == nil {
		h.opts.Level = slog.LevelInfo
	}
	h.timeFormat = "2006-01-02 15:04:05.000"
	h.levelMap = map[slog.Level]string{
		slog.LevelDebug: "DBG",
		slog.LevelInfo:  "INF",
		slog.LevelWarn:  "WRN",
		slog.LevelError: "ERR",
	}
	return h
}

func (h *PlainTextHandler) Enabled(ctx context.Context, level slog.Level) bool {
	minLevel := slog.LevelInfo
	if h.opts.Level != nil {
		minLevel = h.opts.Level.Level()
	}
	return level >= minLevel
}

func (h *PlainTextHandler) Handle(ctx context.Context, r slog.Record) error {
	buf := &strings.Builder{}

	if !r.Time.IsZero() {
		buf.WriteString(r.Time.Format(h.timeFormat))
		buf.WriteString(" ")
	}

	if levelStr, ok := h.levelMap[r.Level]; ok {
		buf.WriteString(levelStr)
		buf.WriteString(" ")
	} else {
		buf.WriteString(r.Level.String())
		buf.WriteString(" ")
	}

	if h.opts.AddSource && r.PC != 0 {
		fs := runtime.CallersFrames([]uintptr{r.PC})
		f, _ := fs.Next()
		if f.File != "" {
			buf.WriteString(fmt.Sprintf("%s:%d ", filepath.Base(f.File), f.Line))
		}
	}

	if h.prefix != "" {
		buf.WriteString("[" + h.prefix + "] ")
	}

	buf.WriteString(r.Message)

	for _, attr := range h.attrs {
		buf.WriteString(" ")
		appendAttr(buf, attr)
	}

	r.Attrs(func(a slog.Attr) bool {
		buf.WriteString(" ")
		appendAttr(buf, a)
		return true
	})

	buf.WriteString("\n")

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := h.w.Write([]byte(buf.String()))
	return err
}

func appendAttr(buf *strings.Builder, a slog.Attr) {
	buf.WriteString(a.Key)
	buf.WriteString("=")
	appendValue(buf, a.Value)
}

func appendValue(buf *strings.Builder, v slog.Value) {
	v = v.Resolve()
	switch v.Kind() {
	case slog.KindString:
		buf.WriteString(strconv.Quote(v.String()))
	case slog.KindTime:
		buf.WriteString(v.Time().Format(time.RFC3339Nano))
	case slog.KindGroup:
		attrs := v.Group()
		if len(attrs) == 0 {
			return
		}
		buf.WriteString("{")
		for i, a := range attrs {
			if i > 0 {
				buf.WriteString(" ")
			}
			appendAttr(buf, a)
		}
		buf.WriteString("}")
	default:
		buf.WriteString(fmt.Sprintf("%v", v.Any()))
	}
}

func (h *PlainTextHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	h2 := *h
	h2.mu = &sync.Mutex{}
	h2.attrs = append(h2.attrs, attrs...)
	return &h2
}

func (h *PlainTextHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	h2 := *h
	h2.mu = &sync.Mutex{}
	if h2.prefix != "" {
		h2.prefix += "."
	}
	h2.prefix += name
	return &h2
}
