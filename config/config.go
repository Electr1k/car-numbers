package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env         string `env:"ENV" env-default:"development"`
	HttpServer  `env-prefix:"HTTP_"`
	LogLevel    string `env:"LOG_LEVEL" env-default:"info"`
	DatabaseURL string `env:"DATABASE_URL" env-required:"true"`
	JwtSecret   string `env:"JWT_SECRET" env-required:"true"`
}

type HttpServer struct {
	Address     string        `env:"ADDRESS" env-default:""`
	Port        string        `env:"PORT" env-default:"8080"`
	Timeout     time.Duration `env:"TIMEOUT" env-default:"5s"`
	IdleTimeout time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
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
