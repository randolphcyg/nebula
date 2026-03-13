package config

import (
	"fmt"
	"os"
	"sync"
	"time"

	"gopkg.in/yaml.v3"
)

var (
	cfg  *Config
	once sync.Once
)

// Config 应用配置
type Config struct {
	App       AppConfig       `yaml:"app"`
	Wireshark WiresharkConfig `yaml:"wireshark"`
	Database  DatabaseConfig  `yaml:"database"`
	Auth      AuthConfig      `yaml:"auth"`
	Log       LogConfig       `yaml:"log"`
	Pcap      PcapConfig      `yaml:"pcap"`
	Server    ServerConfig    `yaml:"server"`
}

// AppConfig 应用配置
type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
	Debug   bool   `yaml:"debug"`
}

// WiresharkConfig Wireshark 服务配置
type WiresharkConfig struct {
	BaseURL            string `yaml:"base_url"`
	ContainerMountPath string `yaml:"container_mount_path"`
	Timeout            int    `yaml:"timeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Type            string `yaml:"type"`
	SQLitePath      string `yaml:"sqlite_path"`
	MaxOpenConns    int    `yaml:"max_open_conns"`
	MaxIdleConns    int    `yaml:"max_idle_conns"`
	ConnMaxLifetime int    `yaml:"conn_max_lifetime"`
	MySQL           MySQLConfig `yaml:"mysql"`
}

// MySQLConfig MySQL 数据库配置
type MySQLConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	SecretKey         string `yaml:"secret_key"`
	TokenExpiry       int    `yaml:"token_expiry"`
	RefreshTokenExpiry int   `yaml:"refresh_token_expiry"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level     string `yaml:"level"`
	Format    string `yaml:"format"`
	FilePath  string `yaml:"file_path"`
	MaxSize   int    `yaml:"max_size"`
	MaxAge    int    `yaml:"max_age"`
}

// PcapConfig PCAP 文件配置
type PcapConfig struct {
	StoragePath       string   `yaml:"storage_path"`
	MaxUploadSize     int      `yaml:"max_upload_size"`
	AllowedExtensions []string `yaml:"allowed_extensions"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Host         string `yaml:"host"`
	Port         int    `yaml:"port"`
	ReadTimeout  int    `yaml:"read_timeout"`
	WriteTimeout int    `yaml:"write_timeout"`
}

// Load 加载配置文件
func Load(configPath string) (*Config, error) {
	var err error
	once.Do(func() {
		cfg, err = loadConfig(configPath)
	})
	return cfg, err
}

// loadConfig 从文件加载配置
func loadConfig(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败：%w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败：%w", err)
	}

	// 设置默认值
	config.setDefaults()

	return &config, nil
}

// setDefaults 设置配置默认值
func (c *Config) setDefaults() {
	if c.App.Name == "" {
		c.App.Name = "NEBULA"
	}
	if c.App.Version == "" {
		c.App.Version = "1.0.0"
	}
	if c.Wireshark.Timeout == 0 {
		c.Wireshark.Timeout = 60
	}
	if c.Database.MaxOpenConns == 0 {
		c.Database.MaxOpenConns = 100
	}
	if c.Database.MaxIdleConns == 0 {
		c.Database.MaxIdleConns = 10
	}
	if c.Database.ConnMaxLifetime == 0 {
		c.Database.ConnMaxLifetime = 3600
	}
	if c.Auth.TokenExpiry == 0 {
		c.Auth.TokenExpiry = 24
	}
	if c.Auth.RefreshTokenExpiry == 0 {
		c.Auth.RefreshTokenExpiry = 168
	}
	if c.Log.Level == "" {
		c.Log.Level = "info"
	}
	if c.Log.Format == "" {
		c.Log.Format = "text"
	}
	if c.Pcap.StoragePath == "" {
		c.Pcap.StoragePath = "./pcaps"
	}
	if c.Pcap.MaxUploadSize == 0 {
		c.Pcap.MaxUploadSize = 500
	}
	if c.Server.Port == 0 {
		c.Server.Port = 8080
	}
	if c.Server.ReadTimeout == 0 {
		c.Server.ReadTimeout = 30
	}
	if c.Server.WriteTimeout == 0 {
		c.Server.WriteTimeout = 30
	}
}

// Get 获取全局配置
func Get() *Config {
	if cfg == nil {
		panic("配置未加载，请先调用 config.Load()")
	}
	return cfg
}

// GetWiresharkTimeout 获取 Wireshark 超时时间（time.Duration）
func (c *WiresharkConfig) GetTimeout() time.Duration {
	return time.Duration(c.Timeout) * time.Second
}

// GetMySQLConnectionString 获取 MySQL 连接字符串
func (d *DatabaseConfig) GetMySQLConnectionString() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		d.MySQL.Username,
		d.MySQL.Password,
		d.MySQL.Host,
		d.MySQL.Port,
		d.MySQL.Database,
	)
}
