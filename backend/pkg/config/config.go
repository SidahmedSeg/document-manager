package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

// Config holds all configuration for the application
type Config struct {
	Environment string         `mapstructure:"ENVIRONMENT"`
	AppName     string         `mapstructure:"APP_NAME"`
	AppVersion  string         `mapstructure:"APP_VERSION"`
	Server      ServerConfig   `mapstructure:",squash"`
	Database    DatabaseConfig `mapstructure:",squash"`
	Redis       RedisConfig    `mapstructure:",squash"`
	Auth        AuthConfig     `mapstructure:",squash"`
	Logger      LoggerConfig   `mapstructure:",squash"`
	Services    ServicesConfig `mapstructure:",squash"`
}

// ServerConfig holds HTTP server configuration
type ServerConfig struct {
	Host         string        `mapstructure:"SERVER_HOST"`
	Port         int           `mapstructure:"SERVER_PORT"`
	ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
	IdleTimeout  time.Duration `mapstructure:"SERVER_IDLE_TIMEOUT"`
}

// DatabaseConfig holds PostgreSQL configuration
type DatabaseConfig struct {
	Host            string        `mapstructure:"DB_HOST"`
	Port            int           `mapstructure:"DB_PORT"`
	User            string        `mapstructure:"DB_USER"`
	Password        string        `mapstructure:"DB_PASSWORD"`
	Name            string        `mapstructure:"DB_NAME"`
	SSLMode         string        `mapstructure:"DB_SSL_MODE"`
	MaxOpenConns    int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	MaxIdleConns    int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	ConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	Host       string `mapstructure:"REDIS_HOST"`
	Port       int    `mapstructure:"REDIS_PORT"`
	Password   string `mapstructure:"REDIS_PASSWORD"`
	DB         int    `mapstructure:"REDIS_DB"`
	MaxRetries int    `mapstructure:"REDIS_MAX_RETRIES"`
	PoolSize   int    `mapstructure:"REDIS_POOL_SIZE"`
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	KratosPublicURL  string `mapstructure:"SHARED_KRATOS_PUBLIC_URL"`
	KratosAdminURL   string `mapstructure:"SHARED_KRATOS_ADMIN_URL"`
	HydraPublicURL   string `mapstructure:"SHARED_HYDRA_PUBLIC_URL"`
	HydraAdminURL    string `mapstructure:"SHARED_HYDRA_ADMIN_URL"`
	OAuth2ClientID   string `mapstructure:"OAUTH2_CLIENT_ID"`
	OAuth2ClientSecret string `mapstructure:"OAUTH2_CLIENT_SECRET"`
	JWTIssuer        string `mapstructure:"JWT_ISSUER"`
	JWTAudience      string `mapstructure:"JWT_AUDIENCE"`
	HydraJWKSURL     string `mapstructure:"HYDRA_JWKS_URL"`
	InternalAPISecret string `mapstructure:"INTERNAL_API_SECRET"`
}

// LoggerConfig holds logging configuration
type LoggerConfig struct {
	Level  string `mapstructure:"LOG_LEVEL"`
	Format string `mapstructure:"LOG_FORMAT"`
}

// ServicesConfig holds microservice URLs
type ServicesConfig struct {
	TenantServiceURL        string `mapstructure:"TENANT_SERVICE_URL"`
	DocumentServiceURL      string `mapstructure:"DOCUMENT_SERVICE_URL"`
	StorageServiceURL       string `mapstructure:"STORAGE_SERVICE_URL"`
	ShareServiceURL         string `mapstructure:"SHARE_SERVICE_URL"`
	RBACServiceURL          string `mapstructure:"RBAC_SERVICE_URL"`
	QuotaServiceURL         string `mapstructure:"QUOTA_SERVICE_URL"`
	OCRServiceURL           string `mapstructure:"OCR_SERVICE_URL"`
	CategorizationServiceURL string `mapstructure:"CATEGORIZATION_SERVICE_URL"`
	SearchServiceURL        string `mapstructure:"SEARCH_SERVICE_URL"`
	NotificationServiceURL  string `mapstructure:"NOTIFICATION_SERVICE_URL"`
	AuditServiceURL         string `mapstructure:"AUDIT_SERVICE_URL"`
}

// GetDSN returns the PostgreSQL connection string
func (c *DatabaseConfig) GetDSN() string {
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		c.Host, c.Port, c.User, c.Password, c.Name, c.SSLMode,
	)
}

// GetRedisAddr returns the Redis address
func (c *RedisConfig) GetRedisAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// GetServerAddr returns the server address
func (c *ServerConfig) GetServerAddr() string {
	return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

// IsDevelopment checks if running in development mode
func (c *Config) IsDevelopment() bool {
	return c.Environment == "development"
}

// IsProduction checks if running in production mode
func (c *Config) IsProduction() bool {
	return c.Environment == "production"
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Environment variables
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// LoadFromFile loads configuration from a file
func LoadFromFile(configPath string) (*Config, error) {
	v := viper.New()

	// Set config file
	v.SetConfigFile(configPath)

	// Read config file
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	// Environment variables override file config
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Set defaults
	setDefaults(v)

	// Unmarshal into Config struct
	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Validate configuration
	if err := validate(&cfg); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// setDefaults sets default configuration values
func setDefaults(v *viper.Viper) {
	// Environment
	v.SetDefault("ENVIRONMENT", "development")
	v.SetDefault("APP_NAME", "Document Manager")
	v.SetDefault("APP_VERSION", "1.0.0")

	// Server
	v.SetDefault("SERVER_HOST", "0.0.0.0")
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_READ_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_WRITE_TIMEOUT", 30*time.Second)
	v.SetDefault("SERVER_IDLE_TIMEOUT", 120*time.Second)

	// Database
	v.SetDefault("DB_HOST", "localhost")
	v.SetDefault("DB_PORT", 15432)
	v.SetDefault("DB_USER", "postgres")
	v.SetDefault("DB_NAME", "docmanager")
	v.SetDefault("DB_SSL_MODE", "disable")
	v.SetDefault("DB_MAX_OPEN_CONNS", 100)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", 3600*time.Second)

	// Redis
	v.SetDefault("REDIS_HOST", "localhost")
	v.SetDefault("REDIS_PORT", 16379)
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("REDIS_MAX_RETRIES", 3)
	v.SetDefault("REDIS_POOL_SIZE", 10)

	// Logger
	v.SetDefault("LOG_LEVEL", "info")
	v.SetDefault("LOG_FORMAT", "json")

	// Auth
	v.SetDefault("JWT_ISSUER", "http://shared-hydra:14444")
	v.SetDefault("JWT_AUDIENCE", "document-manager-client")
}

// validate validates the configuration
func validate(cfg *Config) error {
	// Required fields
	if cfg.Database.Password == "" {
		return fmt.Errorf("DB_PASSWORD is required")
	}

	if cfg.Redis.Password == "" {
		return fmt.Errorf("REDIS_PASSWORD is required")
	}

	if cfg.Auth.InternalAPISecret == "" {
		return fmt.Errorf("INTERNAL_API_SECRET is required")
	}

	// Validate environment
	validEnvs := []string{"development", "staging", "production"}
	isValid := false
	for _, env := range validEnvs {
		if cfg.Environment == env {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("ENVIRONMENT must be one of: %v", validEnvs)
	}

	// Validate log level
	validLevels := []string{"debug", "info", "warn", "error"}
	isValid = false
	for _, level := range validLevels {
		if cfg.Logger.Level == level {
			isValid = true
			break
		}
	}
	if !isValid {
		return fmt.Errorf("LOG_LEVEL must be one of: %v", validLevels)
	}

	return nil
}
