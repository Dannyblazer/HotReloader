package main

import (
	"fmt"
	"os"

	"hotreloader/pkg/optimizer"
	"hotreloader/pkg/watcher"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: hotreloader <directory>")
		os.Exit(1)
	}

	dir := os.Args[1]

	// Initialize the optimizer
	opt := optimizer.NewOptimizer()

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
