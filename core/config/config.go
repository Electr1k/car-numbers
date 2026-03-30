package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `env:"ENV" env-default:"development"`
	HttpServer     `env-prefix:"HTTP_"`
	LogConfig      `env-prefix:"LOG_"`
	DatabaseConfig
	JwtSecret      string `env:"JWT_SECRET" env-required:"true"`
	ProviderConfig `env-prefix:"PROVIDER_"`
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
	URL             string        `env:"DATABASE_URL" env-required:"true"`
}

type ProviderConfig struct {
	AutoNomeraBaseURL string        `env:"AUTONOMERA_BASE_URL" env-default:"https://autonomera777.net/ajax/"`
	AutoNomeraTimeout time.Duration `env:"AUTONOMERA_TIMEOUT" env-default:"60s"`
	BatchSize         int           `env:"BATCH_SIZE" env-default:"20"`
	StartPosition     int           `env:"START_POSITION" env-default:"0"`
	StopAfter         time.Duration `env:"STOP_AFTER" env-default:"72h"`
	RateLimitDelay    time.Duration `env:"RATE_LIMIT_DELAY" env-default:"1s"`
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
