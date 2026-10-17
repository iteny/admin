// Package config 负责加载和校验 CMS 的运行配置。
package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config 是应用的全部配置。
type Config struct {
	App      AppConfig      `mapstructure:"app"`
	Server   ServerConfig   `mapstructure:"server"`
	Database DatabaseConfig `mapstructure:"database"`
	Session  SessionConfig  `mapstructure:"session"`
	Log      LogConfig      `mapstructure:"log"`
	Mail     MailConfig     `mapstructure:"mail"`
}

// AppConfig 是站点级配置。
type AppConfig struct {
	Name             string   `mapstructure:"name"`
	Env              string   `mapstructure:"env"`
	Debug            bool     `mapstructure:"debug"`
	AutoMigrate      bool     `mapstructure:"auto_migrate"`
	Seed             bool     `mapstructure:"seed"`
	Timezone         string   `mapstructure:"timezone"`
	BaseURL          string   `mapstructure:"base_url"`
	UploadMaxSize    int64    `mapstructure:"upload_max_size"`
	UploadAllowedExt []string `mapstructure:"upload_allowed_ext"`
}

// ServerConfig 是 HTTP 服务配置。
type ServerConfig struct {
	Addr            string        `mapstructure:"addr"`
	ReadTimeout     time.Duration `mapstructure:"read_timeout"`
	WriteTimeout    time.Duration `mapstructure:"write_timeout"`
	IdleTimeout     time.Duration `mapstructure:"idle_timeout"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
}

// DatabaseConfig 是数据库连接配置。
type DatabaseConfig struct {
	Driver      string        `mapstructure:"driver"`
	DSN         string        `mapstructure:"dsn"`
	MaxOpen     int           `mapstructure:"max_open"`
	MaxIdle     int           `mapstructure:"max_idle"`
	MaxLifetime time.Duration `mapstructure:"max_lifetime"`
}

// SessionConfig 是会话配置。
type SessionConfig struct {
	Name   string `mapstructure:"name"`
	Secret string `mapstructure:"secret"`
	Secure bool   `mapstructure:"secure"`
}

// LogConfig 是日志配置。
type LogConfig struct {
	Level string `mapstructure:"level"`
	Dir   string `mapstructure:"dir"`
	File  string `mapstructure:"file"`
}

// MailConfig 是邮件配置（当前版本只保留配置项）。
type MailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Username string `mapstructure:"username"`
	Password string `mapstructure:"password"`
	From     string `mapstructure:"from"`
}

// Load 从 YAML 文件加载配置，并允许 CMS_ 前缀的环境变量覆盖。
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")
	v.SetEnvPrefix("CMS")
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	v.AutomaticEnv()
	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}
	cfg.normalize()
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("app.name", "Tabler CMS")
	v.SetDefault("app.env", "development")
	v.SetDefault("app.debug", true)
	v.SetDefault("app.auto_migrate", true)
	v.SetDefault("app.seed", true)
	v.SetDefault("app.timezone", "Asia/Shanghai")
	v.SetDefault("app.base_url", "http://localhost:8080")
	v.SetDefault("app.upload_max_size", 10<<20)
	v.SetDefault("app.upload_allowed_ext", []string{"jpg", "jpeg", "png", "gif", "webp"})

	v.SetDefault("server.addr", ":8080")
	v.SetDefault("server.read_timeout", 15*time.Second)
	v.SetDefault("server.write_timeout", 30*time.Second)
	v.SetDefault("server.idle_timeout", 60*time.Second)
	v.SetDefault("server.shutdown_timeout", 15*time.Second)

	v.SetDefault("database.driver", "sqlite")
	v.SetDefault("database.dsn", "storage/cms.db")
	v.SetDefault("database.max_open", 50)
	v.SetDefault("database.max_idle", 10)
	v.SetDefault("database.max_lifetime", time.Hour)

	v.SetDefault("session.name", "cms_session")
	v.SetDefault("session.secret", "change-this-session-secret-in-production")
	v.SetDefault("session.secure", false)

	v.SetDefault("log.level", "debug")
	v.SetDefault("log.dir", "storage/logs")
	v.SetDefault("log.file", "cms.log")
}

func (c *Config) normalize() {
	if c.App.Env == "" {
		c.App.Env = "development"
	}
	if c.Server.Addr == "" {
		c.Server.Addr = ":8080"
	}
	if c.Database.Driver == "" {
		c.Database.Driver = "sqlite"
	}
	if c.Database.MaxOpen == 0 {
		c.Database.MaxOpen = 50
	}
	if c.Database.MaxIdle == 0 {
		c.Database.MaxIdle = 10
	}
	if c.Database.MaxLifetime == 0 {
		c.Database.MaxLifetime = time.Hour
	}
	if c.Server.ShutdownTimeout == 0 {
		c.Server.ShutdownTimeout = 15 * time.Second
	}
	if c.Log.Dir == "" {
		c.Log.Dir = "storage/logs"
	}
	if c.Log.File == "" {
		c.Log.File = "cms.log"
	}
}

func (c *Config) validate() error {
	switch c.Database.Driver {
	case "sqlite", "mysql":
	default:
		return fmt.Errorf("unsupported database driver %q (want sqlite or mysql)", c.Database.Driver)
	}
	if c.Database.DSN == "" {
		return fmt.Errorf("database.dsn is required")
	}
	if len(c.Session.Secret) < 16 {
		return fmt.Errorf("session.secret must be at least 16 characters")
	}
	if c.IsProduction() && strings.Contains(c.Session.Secret, "change-this") {
		return fmt.Errorf("session.secret must be changed in production")
	}
	if c.App.Timezone == "" {
		c.App.Timezone = "Asia/Shanghai"
	}
	return nil
}

// IsProduction 报告当前是否为生产环境。
func (c *Config) IsProduction() bool {
	return strings.EqualFold(c.App.Env, "production")
}

// UploadAllowed 判断扩展名是否在上传白名单中。
func (c *Config) UploadAllowed(ext string) bool {
	ext = strings.TrimPrefix(strings.ToLower(ext), ".")
	for _, allowed := range c.App.UploadAllowedExt {
		if strings.TrimPrefix(strings.ToLower(allowed), ".") == ext {
			return true
		}
	}
	return false
}

// SessionSecret 返回会话签名密钥。
func (c *Config) SessionSecret() []byte {
	return []byte(c.Session.Secret)
}
