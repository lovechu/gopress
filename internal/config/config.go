package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

// Config 根配置结构
type Config struct {
	App         AppConfig
	Database    DatabaseConfig
	Redis       RedisConfig
	JWT         JWTConfig
	Storage     StorageConfig
	Upload      UploadConfig
	Media       MediaConfig
	Cache       CacheConfig
	Log         LogConfig
	CORS        CORSConfig
	RateLimit   RateLimitConfig
	Gzip        GzipConfig
	Performance PerformanceConfig
}

// ---- 子配置段 ----

type AppConfig struct {
	Name      string
	Env       string
	Port      string
	BaseURL   string
	BasePath  string
	Debug     bool
	Timezone  string
	SecretKey string
}

type DatabaseConfig struct {
	Driver          string
	Host            string
	Port            int
	Name            string
	User            string
	Password        string
	Charset         string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int  // seconds
	ConnMaxIdleTime int  // seconds
	LogLevel        string
}

type RedisConfig struct {
	Addr            string
	Password        string
	DB              int
	PoolSize        int
	MinIdleConns    int
	MaxRetries      int
	DialTimeout     int // seconds
	ReadTimeout     int // seconds
	WriteTimeout    int // seconds
	Prefix          string
}

type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpire  int
	RefreshExpire int
}

type StorageConfig struct {
	Driver string
	Local  LocalStorageConfig
}

type LocalStorageConfig struct {
	Root    string
	BaseURL string
}

type UploadConfig struct {
	MaxSize      int64
	AllowedTypes []string
	ImageQuality int
	GenerateWebP bool
}

type MediaConfig struct {
	Storage      string
	MaxFileSize  int64
	AllowedTypes []string
	Local        LocalMediaConfig
	MinIO        MinIOMediaConfig
}

type LocalMediaConfig struct {
	BasePath string
	BaseURL  string
}

type MinIOMediaConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	Region    string
	BaseURL   string
}

type CacheConfig struct {
	Driver     string
	DefaultTTL int
}

type LogConfig struct {
	Level  string
	Format string
	Output string
}

type CORSConfig struct {
	AllowedOrigins []string
	AllowedMethods []string
	AllowedHeaders []string
	MaxAge        int
}

type RateLimitConfig struct {
	Enabled  bool
	Requests int
	Window   int
}

type GzipConfig struct {
	Enabled         bool
	CompressionLevel int
	MinSize         int // bytes
}

type PerformanceConfig struct {
	Enabled            bool
	ReportInterval     int // seconds
	SlowRequestThreshold int // milliseconds
}

// Load 加载 config.yaml
func Load(path string) (*Config, error) {
	v := viper.New()
	v.SetConfigFile(path)
	v.SetConfigType("yaml")

	// 环境变量覆盖（GOPRESS_ 开头的变量）
	v.SetEnvPrefix("GOPRESS")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}
