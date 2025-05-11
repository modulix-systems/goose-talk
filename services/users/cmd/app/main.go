package main

import (
	"log"

	"github.com/AlexeyTarasov77/talka-chats/internal/app"
	"github.com/AlexeyTarasov77/talka-chats/internal/config"
)

func main() {
	// Configuration
	cfg, err := config.NewConfig()
	if err != nil {
		log.Fatalf("Config error: %s", err)
	}

	// Run
	app.Run(cfg)
}
