package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env              string `env:"ENV" env-default:"development"`
	HttpServer       `env-prefix:"HTTP_"`
	LogConfig        `env-prefix:"LOG_"`
	DatabaseConfig   `env-prefix:"DATABASE_"`
	AutoNomeraConfig `env-prefix:"AUTONOMERA_"`
}

type HttpServer struct {
	Address     string        `env:"ADDRESS" env-default:""`
	Port        string        `env:"PORT" env-default:"8080"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
}

type LogConfig struct {
	Level  string `env:"LEVEL" env-default:"debug"`
	Format string `env:"FORMAT" env-default:"json"`
}

type DatabaseConfig struct {
	URL string `env:"URL" env-required:"true"`
}

// AutoNomeraConfig - Конфиг для autonomera777
type AutoNomeraConfig struct {
	BaseURL   string `env:"BASE_URL" env-default:"https://autonomera777.ru"`
	BatchSize int    `env:"BATCH_SIZE" env-default:"20"`
}

func MustLoad() *Config {
	if err := godotenv.Load(); err != nil {
		log.Printf("No .env file found, using environment variables")
	}

	var cfg Config

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		log.Fatalf("cannot read config: %s", err)
	}

	return &cfg
}
