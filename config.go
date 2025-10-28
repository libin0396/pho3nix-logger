package pho3nix_logger

import (
	"github.com/libin0396/quant-platform-sdk/config"
	"log/slog"
	"os"
)

// RotationConfig 定义了文件轮转的参数
type RotationConfig struct {
	MaxSizeMB  int  `mapstructure:"maxSizeMB"`
	MaxBackups int  `mapstructure:"maxBackups"`
	MaxAgeDays int  `mapstructure:"maxAgeDays"`
	Compress   bool `mapstructure:"compress"`
}

// FileOutputConfig 定义了单个日志级别的文件输出
type FileOutputConfig struct {
	Enabled    bool   `mapstructure:"enabled"`
	Path       string `mapstructure:"path"`
	MaxSizeMB  int    `mapstructure:"maxSizeMB"`
	MaxBackups int    `mapstructure:"maxBackups"`
	MaxAgeDays int    `mapstructure:"maxAgeDays"`
	Compress   bool   `mapstructure:"compress"`
}

// Config 是整个日志系统的配置结构体
type Config struct {
	Level     string `mapstructure:"level"`
	AddSource bool   `mapstructure:"addSource"`
	Console   struct {
		Enabled bool   `mapstructure:"enabled"`
		Level   string `mapstructure:"level"`
	} `mapstructure:"console"`
	File struct {
		DefaultRotation RotationConfig   `mapstructure:"defaultRotation"`
		Debug           FileOutputConfig `mapstructure:"debug"`
		Info            FileOutputConfig `mapstructure:"info"`
		Warn            FileOutputConfig `mapstructure:"warn"`
		Error           FileOutputConfig `mapstructure:"error"`
	} `mapstructure:"file"`
}

func InitializeFromConfig(configKey string) {
	// 1. 直接从 config-sdk 获取 viper 实例
	v := config.GetViper()

	// 2. 从 viper 实例中解析出 logger 的配置节
	var logCfg Config
	if err := v.UnmarshalKey(configKey, &logCfg); err != nil {
		// 如果解析失败，使用最原始的方式打印错误并退出
		slog.Error("加载日志配置失败，程序退出", "config_key", configKey, "error", err)
		os.Exit(1)
	}

	Initialize(&logCfg)
}
