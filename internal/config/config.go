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
	App        AppConfig        `yaml:"app"`
	Wireshark  WiresharkConfig  `yaml:"wireshark"`
	Database   DatabaseConfig   `yaml:"database"`
	Auth       AuthConfig       `yaml:"auth"`
	Log        LogConfig        `yaml:"log"`
	Pcap       PcapConfig       `yaml:"pcap"`
	Server     ServerConfig     `yaml:"server"`
	ZeekRunner ZeekRunnerConfig `yaml:"zeek_runner"`
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
	Type            string      `yaml:"type"`
	SQLitePath      string      `yaml:"sqlite_path"`
	MaxOpenConns    int         `yaml:"max_open_conns"`
	MaxIdleConns    int         `yaml:"max_idle_conns"`
	ConnMaxLifetime int         `yaml:"conn_max_lifetime"`
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
	SecretKey          string `yaml:"secret_key"`
	TokenExpiry        int    `yaml:"token_expiry"`
	RefreshTokenExpiry int    `yaml:"refresh_token_expiry"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level    string `yaml:"level"`
	Format   string `yaml:"format"`
	FilePath string `yaml:"file_path"`
	MaxSize  int    `yaml:"max_size"`
	MaxAge   int    `yaml:"max_age"`
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

// ZeekRunnerConfig Zeek Runner 服务配置
type ZeekRunnerConfig struct {
	GRPCAddress string          `yaml:"grpc_address"`
	HTTPAddress string          `yaml:"http_address"`
	Timeout     int             `yaml:"timeout"`
	Enabled     bool            `yaml:"enabled"`
	Retry       ZeekRetryConfig `yaml:"retry"`
	Pool        ZeekPoolConfig  `yaml:"pool"`
}

// ZeekRetryConfig Zeek 重试配置
type ZeekRetryConfig struct {
	MaxAttempts      int `yaml:"max_attempts"`
	BackoffBaseDelay int `yaml:"backoff_base_delay"`
	BackoffMaxDelay  int `yaml:"backoff_max_delay"`
}

// ZeekPoolConfig Zeek 连接池配置
type ZeekPoolConfig struct {
	MaxIdleConns    int `yaml:"max_idle_conns"`
	MaxOpenConns    int `yaml:"max_open_conns"`
	ConnMaxLifetime int `yaml:"conn_max_lifetime"`
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
	if c.ZeekRunner.GRPCAddress == "" {
		c.ZeekRunner.GRPCAddress = "localhost:50051"
	}
	if c.ZeekRunner.HTTPAddress == "" {
		c.ZeekRunner.HTTPAddress = "http://localhost:8080"
	}
	if c.ZeekRunner.Timeout == 0 {
		c.ZeekRunner.Timeout = 30
	}
	if c.ZeekRunner.Retry.MaxAttempts == 0 {
		c.ZeekRunner.Retry.MaxAttempts = 3
	}
	if c.ZeekRunner.Retry.BackoffBaseDelay == 0 {
		c.ZeekRunner.Retry.BackoffBaseDelay = 1
	}
	if c.ZeekRunner.Retry.BackoffMaxDelay == 0 {
		c.ZeekRunner.Retry.BackoffMaxDelay = 10
	}
	if c.ZeekRunner.Pool.MaxIdleConns == 0 {
		c.ZeekRunner.Pool.MaxIdleConns = 10
	}
	if c.ZeekRunner.Pool.MaxOpenConns == 0 {
		c.ZeekRunner.Pool.MaxOpenConns = 20
	}
	if c.ZeekRunner.Pool.ConnMaxLifetime == 0 {
		c.ZeekRunner.Pool.ConnMaxLifetime = 1800
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

// GetZeekTimeout 获取 Zeek 超时时间（time.Duration）
func (z *ZeekRunnerConfig) GetTimeout() time.Duration {
	return time.Duration(z.Timeout) * time.Second
}

// IsEnabled 检查 Zeek 服务是否启用
func (z *ZeekRunnerConfig) IsEnabled() bool {
	return z.Enabled
}

// GetGRPCAddress 获取 gRPC 地址
func (z *ZeekRunnerConfig) GetGRPCAddress() string {
	if z.GRPCAddress == "" {
		return "localhost:50051"
	}
	return z.GRPCAddress
}

// GetHTTPAddress 获取 HTTP 地址
func (z *ZeekRunnerConfig) GetHTTPAddress() string {
	if z.HTTPAddress == "" {
		return "http://localhost:8080"
	}
	return z.HTTPAddress
}
