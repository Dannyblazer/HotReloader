package watcher

import (
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"hotreloader/pkg/optimizer"

	"github.com/fsnotify/fsnotify"
)

// Watcher watches files for changes
type Watcher struct {
	watcher   *fsnotify.Watcher
	optimizer *optimizer.Optimizer
	rootDir   string
	debounce  time.Duration
	ignore    []string
}

// NewWatcher creates a new file watcher
func NewWatcher(rootDir string, opt *optimizer.Optimizer) (*Watcher, error) {
	fsWatcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("failed to create watcher: %w", err)
	}

	w := &Watcher{
		watcher:   fsWatcher,
		optimizer: opt,
		rootDir:   rootDir,
		debounce:  100 * time.Millisecond,
		ignore: []string{
			"node_modules",
			".git",
			".vscode",
			".idea",
			"dist",
			"build",
			"*.log",
			".DS_Store",
		},
	}

	return w, nil
}

// Start begins watching for file changes
func (w *Watcher) Start() error {
	// Add root directory and all subdirectories
	if err := w.addRecursive(w.rootDir); err != nil {
		return err
	}

	// Create a debounce map to prevent rapid repeated events
	debounceMap := make(map[string]time.Time)

	// Handle interrupt signal for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	// Create ticker for periodic stats display
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	fmt.Println("\nWatching for changes... (Press Ctrl+C to show stats and exit)\n")

	for {
		select {
		case event, ok := <-w.watcher.Events:
			if !ok {
				return nil
			}

			// Ignore certain operations
			if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				continue
			}

			// Check if file should be ignored
			if w.shouldIgnore(event.Name) {
				continue
			}

			// Debounce rapid events for the same file
			now := time.Now()
			if lastTime, exists := debounceMap[event.Name]; exists {
				if now.Sub(lastTime) < w.debounce {
					continue
				}
			}
			debounceMap[event.Name] = now

			// Process the change
			if event.Op&fsnotify.Write == fsnotify.Write {
				if err := w.optimizer.ProcessFileChange(event.Name); err != nil {
					fmt.Printf("Error processing %s: %v\n", event.Name, err)
				}
			} else if event.Op&fsnotify.Create == fsnotify.Create {
				// If a directory was created, add it to the watcher
				if info, err := os.Stat(event.Name); err == nil && info.IsDir() {
					w.addRecursive(event.Name)
				} else {
					if err := w.optimizer.ProcessFileChange(event.Name); err != nil {
						fmt.Printf("Error processing %s: %v\n", event.Name, err)
					}
				}
			} else if event.Op&fsnotify.Remove == fsnotify.Remove {
				fmt.Printf("Removed: %s\n", event.Name)
			}

		case err, ok := <-w.watcher.Errors:
			if !ok {
				return nil
			}
			fmt.Printf("Watcher error: %v\n", err)

		case <-ticker.C:
			// Periodically show summary
			w.optimizer.GetDashboard().PrintSummary()

		case <-sigChan:
			fmt.Println("\n\nShutting down...")
			w.optimizer.Shutdown() // Stop running process
			w.optimizer.PrintStats()
			w.optimizer.GetDashboard().PrintSummary()
			return nil
		}
	}
}

// Close stops the watcher
func (w *Watcher) Close() error {
	return w.watcher.Close()
}

// addRecursive adds a directory and all its subdirectories to the watcher
func (w *Watcher) addRecursive(path string) error {
	return filepath.Walk(path, func(walkPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if info.IsDir() {
			// Check if directory should be ignored
			if w.shouldIgnore(walkPath) {
				return filepath.SkipDir
			}

			if err := w.watcher.Add(walkPath); err != nil {
				return fmt.Errorf("failed to add %s: %w", walkPath, err)
			}
		}

		return nil
	})
}

// shouldIgnore checks if a path should be ignored
func (w *Watcher) shouldIgnore(path string) bool {
	base := filepath.Base(path)

	for _, pattern := range w.ignore {
		// Simple pattern matching
		if strings.Contains(pattern, "*") {
			// Handle wildcard patterns
			if matched, _ := filepath.Match(pattern, base); matched {
				return true
			}
		} else if base == pattern || strings.Contains(path, "/"+pattern+"/") {
			return true
		}
	}

	return false
}
