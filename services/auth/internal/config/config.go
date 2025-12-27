package config

import (
	"fmt"
	"os"
	"path"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/modulix-systems/goose-talk/internal/utils"
	"github.com/modulix-systems/goose-talk/pkg/logger"
)

type (
	// Config -.
	Config struct {
		PG      PG
		Log     Log
		Server  Server
	}

	Server struct {
		Port     string
		Hostname string
	}

	// PG -.
	PG struct {
		Dsn     string `env:"PG_URL,required"`
		MaxPoolSize int `env:"PG_MAX_POOL_SIZE" env-default:"10"`
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
		mode = "local"
		fmt.Println("MODE is not defined, use local by default")
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
