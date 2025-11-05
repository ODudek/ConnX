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

    // Start config watcher for hot reload
    watcher, err := config.NewWatcher(*configPath)
    if err != nil {
        log.Printf("Warning: Failed to create config watcher: %v", err)
        log.Println("Hot reload disabled")
    } else {
        watcher.Start()
        defer watcher.Stop()
        log.Println("Hot reload enabled - config changes will be applied automatically")
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

    // Listen for config reloads
    if watcher != nil {
        go func() {
            for newCfg := range watcher.ReloadChannel() {
                if err := srv.Reload(newCfg); err != nil {
                    log.Printf("Error reloading config: %v", err)
                }
            }
        }()
    }

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
