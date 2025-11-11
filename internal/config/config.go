package config

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ory/viper"
)

type HttpServer struct {
	Addr string `yaml:"address" env:"HTTP_ADDRESS" env-required:"true"`
}

type Postgres struct {
	Host     string `yaml:"host" env:"PG_HOST" env-required:"true"`
	Port     int    `yaml:"port" env:"PG_PORT" env-required:"true"`
	User     string `yaml:"user" env:"PG_USER" env-required:"true"`
	Password string `yaml:"password" env:"PG_PASSWORD" env-required:"true"`
	DBName   string `yaml:"dbname" env:"PG_DBNAME" env-required:"true"`
	SSLMode  string `yaml:"sslmode" env:"PG_SSLMODE" env-default:"disable"`
}

type Config struct {
	Env                string     `mapstructure:"env"`
	CORSAllowedOrigins string     `mapstructure:"cors_allowed_origins"`
	DBType             string     `mapstructure:"db_type"`
	HTTPServer         HttpServer `mapstructure:"http_server"`
	Postgres           Postgres   `mapstructure:"postgres"`
	MaxIdleConns       int        `mapstructure:"max_idle_conns"`
	ConnMaxLifetime    int        `mapstructure:"conn_max_lifetime"`
	ConnMaxIdleTime    int        `mapstructure:"conn_max_idle_time"`
}

func MustLoad() *Config {

	v := viper.New()
	v.SetConfigName("local")
	v.SetConfigType("yaml")
	v.AddConfigPath("./config")

	var cfg Config
	configPath := os.Getenv("CONFIG_PATH")
	flag.StringVar(&configPath, "config", "./config/local.yaml", "Path to configuration file")
	flag.Parse()

	if configPath == "" {
		flags := flag.String("config", "", "Path to configuration file")
		flag.Parse()

		configPath = *flags

		if configPath == "" {
			log.Fatal("Config Path not available")
		}
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exits: %s", configPath)
	}

	if err := v.ReadInConfig(); err != nil {
		log.Fatalf("❌ Error reading config file: %v", err)
	}

	if err := v.Unmarshal(&cfg); err != nil {
		log.Fatalf("❌ Unable to decode config: %v", err)
	}

	fmt.Printf("✅ Config loaded for %s\n", cfg.Env)
	return &cfg

}
