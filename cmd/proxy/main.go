package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/ODudek/ConnX/internal/config"
	"github.com/ODudek/ConnX/internal/server"
)

func main() {
    configPath := flag.String("config", "configs/config.yaml", "path to config file")
    flag.Parse()

    cfg, err := config.LoadWithLegacySupport(*configPath)
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

    // Setup signal handling for graceful shutdown
    sigChan := make(chan os.Signal, 1)
    signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

    // Run server in a goroutine
    errChan := make(chan error, 1)
    go func() {
        if err := srv.Run(); err != nil {
            errChan <- err
        }
    }()

    // Wait for shutdown signal or error
    select {
    case sig := <-sigChan:
        log.Printf("Received signal: %v", sig)
        if err := srv.Shutdown(); err != nil {
            log.Printf("Error during shutdown: %v", err)
        }
    case err := <-errChan:
        log.Fatalf("Server error: %v", err)
    }
}
