package config

import (
	"log"
	"os"
	"path/filepath"
	"time"
)

// Watcher monitors configuration file for changes
type Watcher struct {
	configPath   string
	lastModTime  time.Time
	reloadChan   chan *Config
	stopChan     chan bool
	checkInterval time.Duration
}

// NewWatcher creates a new config watcher
func NewWatcher(configPath string) (*Watcher, error) {
	absPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, err
	}

	info, err := os.Stat(absPath)
	if err != nil {
		return nil, err
	}

	return &Watcher{
		configPath:   absPath,
		lastModTime:  info.ModTime(),
		reloadChan:   make(chan *Config, 1),
		stopChan:     make(chan bool),
		checkInterval: 2 * time.Second, // Check every 2 seconds
	}, nil
}

// Start begins watching the config file
func (w *Watcher) Start() {
	go w.watch()
}

// Stop stops the watcher
func (w *Watcher) Stop() {
	close(w.stopChan)
}

// ReloadChannel returns channel that receives new configs
func (w *Watcher) ReloadChannel() <-chan *Config {
	return w.reloadChan
}

// watch monitors the file for changes
func (w *Watcher) watch() {
	ticker := time.NewTicker(w.checkInterval)
	defer ticker.Stop()

	log.Printf("Config watcher started for: %s", w.configPath)

	for {
		select {
		case <-ticker.C:
			w.checkForChanges()
		case <-w.stopChan:
			log.Println("Config watcher stopped")
			return
		}
	}
}

// checkForChanges checks if config file was modified
func (w *Watcher) checkForChanges() {
	info, err := os.Stat(w.configPath)
	if err != nil {
		log.Printf("Error checking config file: %v", err)
		return
	}

	modTime := info.ModTime()
	if modTime.After(w.lastModTime) {
		log.Printf("Config file changed, reloading...")
		w.lastModTime = modTime

		// Small delay to ensure file write is complete
		time.Sleep(100 * time.Millisecond)

		// Load new configuration
		newConfig, err := Load(w.configPath)
		if err != nil {
			log.Printf("Error loading new config: %v", err)
			return
		}

		// Validate new configuration
		if err := newConfig.Validate(); err != nil {
			log.Printf("Invalid new config: %v", err)
			return
		}

		// Send new config through channel (non-blocking)
		select {
		case w.reloadChan <- newConfig:
			log.Println("Configuration reloaded successfully")
		default:
			log.Println("Reload channel full, skipping update")
		}
	}
}
