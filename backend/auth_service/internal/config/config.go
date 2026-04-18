package config

import (
	"log"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env             string        `yaml:"env" env-required:"true"`
	DBPath          string        `yaml:"db_path" env-required:"true"`
	GRPC            GRPCConfig    `yaml:"grpc"`
	TokenAccessTTL  time.Duration `yaml:"token_accessTTL" env-required:"true"`
	TokenRefreshTTL time.Duration `yaml:"token_refreshTTL" env-required:"true"`
	TokenSecret     string        `yaml:"token_secret" env-required:"true"`
	Redis           RedisConfig   `yaml:"redis"`
}

type GRPCConfig struct {
	Port    uint          `yaml:"port" env-required:"true"`
	Timeout time.Duration `yaml:"timeout" env-required:"true"`
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
