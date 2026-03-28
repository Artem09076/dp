package config

import (
	"log"
	"os"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Env  string     `yaml:"env" env-required:"true"`
	SMTP SMTPConfig `yaml:"smtp"`

	RabbitURL string `yaml:"rabbit_url" env-required:"true"`
}

type SMTPConfig struct {
	SMTPHost string `yaml:"smtp_host" env-required:"true"`
	SMTPPort int    `yaml:"smtp_port" env-required:"true"`
	SMTPUser string `yaml:"smtp_user" env-required:"true"`
	SMTPPass string `yaml:"smtp_pass" env-required:"true"`
	SMTPFrom string `yaml:"smtp_from" env-required:"true"`
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
