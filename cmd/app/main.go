package main

import (
	"log"

	"github.com/DealUnloker/mockzoo/config"
	"github.com/DealUnloker/mockzoo/internal/app"
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
