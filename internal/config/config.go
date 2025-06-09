package config

import (
	"fmt"
	"log/slog"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	API    API    `envPrefix:"API_"`
	DB     DB     `envPrefix:"DB_"`
	Logger Logger `envPrefix:"LOGGER_"`
}

type API struct {
	Port int `env:"PORT" envDefault:"8080"`
}

type DB struct {
	Host     string `env:"HOST"`
	Port     string `env:"PORT"`
	User     string `env:"USER"`
	Password string `env:"PASSWORD"`
	Name     string `env:"NAME"`
}

func (d *DB) URL() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", d.User, d.Password, d.Host, d.Port, d.Name)
}

type Logger struct {
	Level         string `env:"LEVEL" envDefault:"info"`
	HumanReadable bool   `env:"HUMAN_READABLE" envDefault:"true"`
}

func (l *Logger) SlogLevel() slog.Level {
	var level slog.Level
	if err := level.UnmarshalText([]byte(l.Level)); err != nil {
		slog.Warn("failed to unmarshal log level.",
			slog.String("level", l.Level),
			slog.Any("error", err),
		)
		return slog.LevelInfo
	}
	return level
}

func Load(filePath *string) (*Config, error) {
	var cfg Config
	if err := env.Parse(&cfg); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &cfg, nil
}
