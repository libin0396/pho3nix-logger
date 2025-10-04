package pho3nix_logger

import "github.com/spf13/viper"

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

// LoadConfigFromViper 从一个 viper 实例中加载日志配置
// v: 任何一个配置了日志参数的 viper 实例
// configKey: 日志配置在 viper 中的顶层键 (e.g., "logger")
func LoadConfigFromViper(v *viper.Viper, configKey string) (*Config, error) {
	var cfg Config
	// 设置一些合理的默认值
	cfg.Level = "info"
	cfg.Console.Enabled = true
	cfg.Console.Level = "debug"
	cfg.File.DefaultRotation = RotationConfig{MaxSizeMB: 50, MaxBackups: 3, MaxAgeDays: 7}

	if err := v.UnmarshalKey(configKey, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
