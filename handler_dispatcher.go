package pho3nix_logger

import (
	"context"
	"log/slog"
)

// LevelDispatcherHandler 负责将日志记录分发到不同级别的 Handler
type LevelDispatcherHandler struct {
	level    slog.Level
	handlers map[slog.Level][]slog.Handler
}

// NewLevelDispatcherHandler 创建一个新的分发器 Handler
func NewLevelDispatcherHandler(level slog.Level, handlers map[slog.Level][]slog.Handler) *LevelDispatcherHandler {
	return &LevelDispatcherHandler{level: level, handlers: handlers}
}

func (h *LevelDispatcherHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *LevelDispatcherHandler) Handle(ctx context.Context, r slog.Record) error {
	if targetHandlers, ok := h.handlers[r.Level]; ok {
		for _, handler := range targetHandlers {
			if handler.Enabled(ctx, r.Level) {
				if err := handler.Handle(ctx, r); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (h *LevelDispatcherHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	newHandlers := make(map[slog.Level][]slog.Handler, len(h.handlers))
	for level, handlerList := range h.handlers {
		newHandlerList := make([]slog.Handler, len(handlerList))
		for i, handler := range handlerList {
			newHandlerList[i] = handler.WithAttrs(attrs)
		}
		newHandlers[level] = newHandlerList
	}
	return NewLevelDispatcherHandler(h.level, newHandlers)
}

func (h *LevelDispatcherHandler) WithGroup(name string) slog.Handler {
	newHandlers := make(map[slog.Level][]slog.Handler, len(h.handlers))
	for level, handlerList := range h.handlers {
		newHandlerList := make([]slog.Handler, len(handlerList))
		for i, handler := range handlerList {
			newHandlerList[i] = handler.WithGroup(name)
		}
		newHandlers[level] = newHandlerList
	}
	return NewLevelDispatcherHandler(h.level, newHandlers)
}
