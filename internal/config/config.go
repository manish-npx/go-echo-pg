package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/spf13/viper"
)

type HTTPServer struct {
	Address string `mapstructure:"address"`
}

type DBConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	User     string `mapstructure:"user"`
	Password string `mapstructure:"password"`
	DBName   string `mapstructure:"dbname"`
	SSLMode  string `mapstructure:"sslmode"`
}

type JWTConfig struct {
	Secret    string `mapstructure:"secret"`
	ExpiresIn int    `mapstructure:"expires_in"`
}

type Config struct {
	Env                string     `mapstructure:"env"`
	HTTPServer         HTTPServer `mapstructure:"http_server"`
	DB                 DBConfig   `mapstructure:"db"`
	JWT                JWTConfig  `mapstructure:"jwt"`
	CORSAllowedOrigins string     `mapstructure:"cors_allowed_origins"`
	DBMaxIdleConns     int        `mapstructure:"db_max_idle_conns"`
	DBConnMaxLifetime  int        `mapstructure:"db_conn_max_lifetime"`
	DBConnMaxIdleTime  int        `mapstructure:"db_conn_max_idle_time"`
}

func MustLoad() *Config {
	var configPath string
	flag.StringVar(&configPath, "config", "./config/local.yaml", "Path to config file")
	flag.Parse()

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist: %s", configPath)
	}

	v := viper.New()
	v.SetConfigFile(configPath)
	v.SetConfigType("yaml")

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}

	fmt.Printf("✅ config loaded (%s)\n", cfg.Env)
	return &cfg
}
