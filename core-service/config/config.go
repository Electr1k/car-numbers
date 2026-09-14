package config

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env            string `env:"ENV" env-default:"development"`
	HTTPServer     `env-prefix:"HTTP_"`
	LogConfig      `env-prefix:"LOG_"`
	DatabaseConfig `env-prefix:"DATABASE_"`
	PlateConfig    `env-prefix:"PLATE_"`
}

type PlateConfig struct {
	URL     string        `env:"URL" env-required:"true"`
	Timeout time.Duration `env:"TIMEOUT" env-default:"10s"`
}

type HTTPServer struct {
	Address         string        `env:"ADDRESS" env-default:""`
	Port            string        `env:"PORT" env-default:"8080"`
	Timeout         time.Duration `env:"TIMEOUT" env-default:"60s"`
	RequestTimeout  time.Duration `env:"REQUEST_TIMEOUT" env-default:"30s"`
	IdleTimeout     time.Duration `env:"IDLE_TIMEOUT" env-default:"60s"`
	ShutdownTimeout time.Duration `env:"SHUTDOWN_TIMEOUT" env-default:"5s"`
}

type LogConfig struct {
	Level  string `env:"LEVEL" env-default:"debug"`
	Format string `env:"FORMAT" env-default:"json"`
}

type DatabaseConfig struct {
	URL            string `env:"URL" env-required:"true"`
	MigrationsPath string `env:"MIGRATIONS_PATH" env-default:"./migrations"`
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
