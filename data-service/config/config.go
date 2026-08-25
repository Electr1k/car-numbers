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
	GosnomeruConfig  `env-prefix:"GOSNOMERU_"`
	CronConfig       `env-prefix:"CRON_"`
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

type AutoNomeraConfig struct {
	BaseURL     string        `env:"BASE_URL" env-default:"https://autonomera777.ru"`
	BatchSize   int           `env:"BATCH_SIZE" env-default:"20"`
	ImportDepth time.Duration `env:"IMPORT_DEPTH" env-default:"72h"`
}

type GosnomeruConfig struct {
	BaseURL     string        `env:"BASE_URL" env-default:"https://gosnomeru.com"`
	ImportDepth time.Duration `env:"IMPORT_DEPTH" env-default:"72h"`
}

type CronConfig struct {
	ImportAutonomeraActiveOffers  string `env:"IMPORT_AUTONOMERA_ACTIVE_OFFERS" env-default:"0 * * * *"`
	ImportAutonomeraArchiveOffers string `env:"IMPORT_AUTONOMERA_ARCHIVE_OFFERS" env-default:"30 * * * *"`
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
