package config

import (
	"fmt"
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/modulix-systems/goose-talk/pkg/logger"
)

type (
	// Config -.
	Config struct {
		PG      PG
		Metrics Metrics
		Log     Log
		Server  Server
	}

	Server struct {
		Port     string
		Hostname string
	}

	// PG -.
	PG struct {
		PoolMax int    `env:"PG_POOL_MAX,required"`
		Dsn     string `env:"PG_URL,required"`
	}

	Metrics struct {
		Enabled bool `env:"METRICS_ENABLED" envDefault:"true"`
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

func ResolveConfigPath() string {
	mode := os.Getenv("MODE")
	if mode == "" {
		panic("Unable to determine config path. MODE env variable is not set")
	}
	currDir, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Failed to build config path. Error: %s", err))
	}
	newPath := path.Join(currDir, "config", mode)
	newPath += ".yaml"
	// try both extensions
	for _, ext := range []string{".yaml", "yml"} {
		if _, err := os.Stat(newPath + ext); err == nil {
			return newPath + ext
		}
	}
	panic(fmt.Sprintf("Failed to build config path. Error: %s", err))
}
