package main

import (
	"flag"

	"github.com/modulix-systems/goose-talk/internal/app"
	"github.com/modulix-systems/goose-talk/internal/config"
)

func main() {
	// Configuration
	var configPath string
	flag.StringVar(&configPath, "config", "./configs/local.yaml", "path to config file")
	flag.Parse()
	if configPath == "" {
		configPath = config.ResolveConfigPath()
	}
	cfg := config.MustLoad(configPath)
	app.Run(cfg)
}
