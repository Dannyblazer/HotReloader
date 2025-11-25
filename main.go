package main

import (
	"fmt"
	"hotreloader/pkg/optimizer"
	"hotreloader/pkg/watcher"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hotreloader <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]

	// Initialize the optimizer with project directory
	opt := optimizer.NewOptimizer(dir)

	// Perform initial build and start the application
	if err := opt.InitialBuild(); err != nil {
		fmt.Fprintf(os.Stderr, "Initial build failed: %v\n", err)
		fmt.Println("Continuing to watch for changes...")
	}

	// Create file watcher
	w, err := watcher.NewWatcher(dir, opt)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating watcher: %v\n", err)
		os.Exit(1)
	}
	defer w.Close()

	fmt.Printf("Hot Reload Optimizer watching: %s\n", dir)
	fmt.Println("Press Ctrl+C to stop...")

	// Start watching
	if err := w.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
