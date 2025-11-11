package config

import (
	"fmt"
	"strings"
	"time"

	"github.com/spf13/viper"
)

type Config struct {
	Env     string `mapstructure:"env"`
	Server  ServerConfig
	DB      DBConfig
	JWT     JWTConfig
	CORS    CORSConfig
	Logging LoggingConfig
}

type ServerConfig struct {
	Address string `mapstructure:"address"`
}

type DBConfig struct {
	Host            string        `mapstructure:"host"`
	Port            int           `mapstructure:"port"`
	User            string        `mapstructure:"user"`
	Password        string        `mapstructure:"password"`
	DBName          string        `mapstructure:"dbname"`
	SSLMode         string        `mapstructure:"sslmode"`
	MaxConns        int           `mapstructure:"max_conns"`
	MinConns        int           `mapstructure:"min_conns"`
	MaxConnLifetime time.Duration `mapstructure:"max_conn_lifetime"`
	MaxConnIdleTime time.Duration `mapstructure:"max_conn_idle_time"`
}

type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

type CORSConfig struct {
	AllowedOrigins string `mapstructure:"allowed_origins"`
}

type LoggingConfig struct {
	Level  string `mapstructure:"level"`
	Format string `mapstructure:"format"`
}

func Load(configPath ...string) (*Config, error) {
	v := viper.New()

	// Set defaults
	setDefaults(v)

	// Bind environment variables
	bindEnvVars(v)

	// If config path is provided, use it
	if len(configPath) > 0 && configPath[0] != "" {
		v.SetConfigFile(configPath[0])
	} else {
		// Default config file names and paths
		v.SetConfigName("config")
		v.SetConfigType("yaml")
		v.AddConfigPath(".")
		v.AddConfigPath("./config")
		v.AddConfigPath("./configs")
	}

	// Read config file (optional - will use defaults if file doesn't exist)
	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			fmt.Printf("Config file not found, using defaults and environment variables\n")
		} else {
			return nil, fmt.Errorf("error reading config: %w", err)
		}
	}

	var config Config
	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, err
	}

	return &config, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("env", "development")
	v.SetDefault("http_server.address", ":8080")
	v.SetDefault("db.host", "localhost")
	v.SetDefault("db.port", 5432)
	v.SetDefault("db.sslmode", "disable")
	v.SetDefault("db.max_conns", 25)
	v.SetDefault("db.min_conns", 5)
	v.SetDefault("db.max_conn_lifetime", time.Hour)
	v.SetDefault("db.max_conn_idle_time", 30*time.Minute)
	v.SetDefault("jwt.expires_in", 3600)
	v.SetDefault("cors.allowed_origins", "http://localhost:3000")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
}

func bindEnvVars(v *viper.Viper) {
	v.SetEnvPrefix("APP")
	v.AutomaticEnv()

	// Map environment variables to config keys
	v.BindEnv("env", "APP_ENV")
	v.BindEnv("http_server.address", "APP_HTTP_SERVER_ADDRESS")
	v.BindEnv("db.host", "APP_DB_HOST")
	v.BindEnv("db.port", "APP_DB_PORT")
	v.BindEnv("db.user", "APP_DB_USER")
	v.BindEnv("db.password", "APP_DB_PASSWORD")
	v.BindEnv("db.dbname", "APP_DB_NAME")
	v.BindEnv("db.sslmode", "APP_DB_SSL_MODE")
	v.BindEnv("jwt.secret", "APP_JWT_SECRET")
	v.BindEnv("cors.allowed_origins", "APP_CORS_ALLOWED_ORIGINS")
	v.BindEnv("logging.level", "APP_LOG_LEVEL")
}

func validateConfig(config *Config) error {
	if config.Env == "production" {
		if config.JWT.Secret == "your-super-secret-key-change-in-production-2025" {
			return fmt.Errorf("JWT secret must be changed in production")
		}
		if config.DB.Password == "" {
			return fmt.Errorf("database password is required in production")
		}
		if !strings.EqualFold(config.DB.SSLMode, "require") && !strings.EqualFold(config.DB.SSLMode, "verify-full") {
			return fmt.Errorf("SSL mode must be 'require' or 'verify-full' in production")
		}
	}
	return nil
}
