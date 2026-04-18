package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env         string        `yaml:"env" env-required:"true"`
	DBPath      string        `yaml:"db_path" env-required:"true"`
	RabbitPath  string        `yaml:"rabbit_url" env-required:"true"`
	HTTP        HTTPConfig    `yaml:"http"`
	TokenTTL    time.Duration `yaml:"token_ttl" env-required:"true"`
	TokenSecret string        `yaml:"token_secret" env-required:"true"`
	Redis       RedisConfig   `yaml:"redis"`
}

type HTTPConfig struct {
	Address     string        `yaml:"address" env-required:"true"`
	Timeout     time.Duration `yaml:"timeout" env-required:"true"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-required:"true"`
}

type RedisConfig struct {
	Addr     string `yaml:"address" env-required:"true"`
	Password string `yaml:"password"`
	DB       int    `yaml:"db"`
}

func LoadConfig() *Config {
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		log.Fatal("CONFIG_PATH environment variable is required")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		log.Fatalf("config file does not exist at path: %s", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("failed to read config: %v", err)
	}
	return &cfg
}
