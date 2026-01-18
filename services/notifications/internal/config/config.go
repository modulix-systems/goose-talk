package config

import (
	"fmt"
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/modulix-systems/goose-talk/internal/utils"
	"github.com/modulix-systems/goose-talk/logger"
)

type (
	// Config -.
	Config struct {
		Postgres Postgres
		RabbitMQ RabbitMQ
		Smtp     Smtp
		Log      Log
		App      App
	}

	App struct {
		Name    string `env:"APP_NAME" env-default:"Goose Talk"`
		Version string `env:"APP_VERSION" env-default:"v0.0.1"`
		Url     string `env:"APP_URL,required"`
	}

	Postgres struct {
		Url         string `env:"PG_URL,required"`
		MaxPoolSize int    `env:"PG_MAX_POOL_SIZE"`
	}

	RabbitMQ struct {
		Url string `env:"RABBIT_URL,required"`
	}

	Smtp struct {
		Host     string `env:"SMTP_HOST,required"`
		Port     string `env:"SMTP_PORT" env-default:"587"`
		Username string `env:"SMTP_USERNAME,required"`
		Password string `env:"SMTP_PASSWORD,required"`
	}

	Log struct {
		Level logger.LogLevel
	}
)

// MustLoad returns app config.
func MustLoad(configPath string) *Config {
	if configPath == "" {
		panic("config path is required")
	}
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config on path " + configPath + " does not exist")
	}
	var cfg Config
	err := cleanenv.ReadConfig(configPath, &cfg)
	if err != nil {
		panic(err)
	}
	return &cfg
}

func ResolveConfigPath(mode string) string {
	if mode == "" {
		mode = os.Getenv("MODE")
		if mode == "" {
			mode = "local"
			fmt.Println("MODE is not defined, use 'local' by default")
		}
	}

	rootPath := utils.FindRootPath()
	newPath := path.Join(rootPath, "configs", mode)
	configExtensions := []string{".yaml", "yml"}

	for _, ext := range configExtensions {
		if _, err := os.Stat(newPath + ext); err == nil {
			return newPath + ext
		}
	}
	panic(fmt.Sprintf("Config for mode '%s' was not found", mode))
}
