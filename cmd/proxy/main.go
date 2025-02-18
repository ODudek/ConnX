package main

import (
	"flag"
	"log"

	"github.com/ODudek/ConnX/internal/config"
	"github.com/ODudek/ConnX/internal/server"
)

func main() {
    configPath := flag.String("config", "configs/config.yaml", "path to config file")
    flag.Parse()

    cfg, err := config.Load(*configPath)
    if err != nil {
        log.Fatalf("Failed to load config: %v", err)
    }

    if err := cfg.Validate(); err != nil {
        log.Fatalf("Invalid config: %v", err)
    }

    srv, err := server.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }

    if err := srv.Run(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
