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
	URL            string `env:"URL" env-required:"true"`
	MigrationsPath string `env:"MIGRATIONS_PATH" env-default:"./migrations"`
}

// AutoNomeraConfig - Конфиг для autonomera777
type AutoNomeraConfig struct {
	BaseURL   string `env:"BASE_URL" env-default:"https://autonomera777.ru"`
	BatchSize int    `env:"BATCH_SIZE" env-default:"20"`

	// Расписание кронов: пауза отсчитывается от конца прогона, а не от начала
	ActiveInterval  time.Duration `env:"ACTIVE_INTERVAL" env-default:"1h"`
	ArchiveInterval time.Duration `env:"ARCHIVE_INTERVAL" env-default:"1h"`

	// ImportDepth - насколько глубоко по дате публикации заходит регулярный
	// прогон. Разовые догрузки задают глубину сами, мимо конфига
	ImportDepth time.Duration `env:"IMPORT_DEPTH" env-default:"72h"`
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
