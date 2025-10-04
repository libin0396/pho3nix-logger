package pho3nix_logger

import (
	"context"
	"gopkg.in/natefinch/lumberjack.v2"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	loggerInitialized sync.Once
	defaultLogger     *slog.Logger
)

func Initialize(cfg *Config) {
	loggerInitialized.Do(func() {
		defaultLogger = setupSlog(cfg)
		slog.SetDefault(defaultLogger)
		slog.Info("Logger SDK initialized successfully")
	})
}

func setupSlog(cfg *Config) *slog.Logger {
	var globalLevel slog.Level
	switch strings.ToLower(cfg.Level) {
	case "debug":
		globalLevel = slog.LevelDebug
	case "info":
		globalLevel = slog.LevelInfo
	case "warn":
		globalLevel = slog.LevelWarn
	case "error":
		globalLevel = slog.LevelError
	default:
		globalLevel = slog.LevelInfo
	}

	handlersMap := make(map[slog.Level][]slog.Handler)
	allLevels := []slog.Level{slog.LevelDebug, slog.LevelInfo, slog.LevelWarn, slog.LevelError}

	if cfg.Console.Enabled {
		var consoleLevel slog.Level
		switch strings.ToLower(cfg.Console.Level) {
		case "debug":
			consoleLevel = slog.LevelDebug
		case "info":
			consoleLevel = slog.LevelInfo
		case "warn":
			consoleLevel = slog.LevelWarn
		case "error":
			consoleLevel = slog.LevelError
		default:
			consoleLevel = slog.LevelDebug
		}

		consoleHandler := NewPlainTextHandler(os.Stderr, &slog.HandlerOptions{
			Level:     consoleLevel,
			AddSource: cfg.AddSource,
		})

		for _, level := range allLevels {
			if level >= consoleLevel {
				handlersMap[level] = append(handlersMap[level], consoleHandler)
			}
		}
	}

	fileConfigs := map[slog.Level]FileOutputConfig{
		slog.LevelDebug: cfg.File.Debug,
		slog.LevelInfo:  cfg.File.Info,
		slog.LevelWarn:  cfg.File.Warn,
		slog.LevelError: cfg.File.Error,
	}

	for level, fileCfg := range fileConfigs {
		if !fileCfg.Enabled {
			continue
		}
		if fileCfg.Path == "" {
			slog.Warn("Log file path is not configured, skipping", "level", level.String())
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fileCfg.Path), 0755); err != nil {
			slog.Error("Failed to create log directory", "path", fileCfg.Path, "error", err)
			continue
		}

		maxSize := fileCfg.MaxSizeMB
		if maxSize == 0 {
			maxSize = cfg.File.DefaultRotation.MaxSizeMB
		}
		maxBackups := fileCfg.MaxBackups
		if maxBackups == 0 {
			maxBackups = cfg.File.DefaultRotation.MaxBackups
		}
		maxAge := fileCfg.MaxAgeDays
		if maxAge == 0 {
			maxAge = cfg.File.DefaultRotation.MaxAgeDays
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   fileCfg.Path,
			MaxSize:    maxSize,
			MaxBackups: maxBackups,
			MaxAge:     maxAge,
			Compress:   fileCfg.Compress,
			LocalTime:  true,
		}

		fileHandler := NewPlainTextHandler(lumberjackLogger, &slog.HandlerOptions{
			Level:     globalLevel,
			AddSource: cfg.AddSource,
		})

		handlersMap[level] = append(handlersMap[level], fileHandler)
	}

	finalHandler := NewLevelDispatcherHandler(globalLevel, handlersMap)
	return slog.New(finalHandler)
}

func Info(msg string, args ...any)  { logWithSourceDepth(1, slog.LevelInfo, msg, args...) }
func Warn(msg string, args ...any)  { logWithSourceDepth(1, slog.LevelWarn, msg, args...) }
func Error(msg string, args ...any) { logWithSourceDepth(1, slog.LevelError, msg, args...) }
func Debug(msg string, args ...any) { logWithSourceDepth(1, slog.LevelDebug, msg, args...) }

func logWithSourceDepth(depth int, level slog.Level, msg string, args ...any) {
	// 确保 logger 总是可用的
	var loggerToUse *slog.Logger
	if defaultLogger != nil {
		loggerToUse = defaultLogger
	} else {
		// 在 Initialize 之前调用的后备 logger
		loggerToUse = slog.New(slog.NewTextHandler(os.Stderr, nil))
	}

	if !loggerToUse.Handler().Enabled(context.Background(), level) {
		return
	}

	var pc uintptr
	var pcs [1]uintptr
	runtime.Callers(3+depth, pcs[:])
	pc = pcs[0]

	r := slog.NewRecord(time.Now(), level, msg, pc)
	r.Add(args...)

	_ = loggerToUse.Handler().Handle(context.Background(), r)
}
