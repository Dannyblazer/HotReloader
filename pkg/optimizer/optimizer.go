package optimizer

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"time"

	"hotreloader/pkg/analyzer"
	"hotreloader/pkg/cache"
	"hotreloader/pkg/dashboard"
	"hotreloader/pkg/plugin"
)

// Optimizer is the core hot reload optimizer
type Optimizer struct {
	cache          *cache.ModuleCache
	analyzer       *analyzer.DependencyAnalyzer
	depGraph       *analyzer.DependencyGraph
	dashboard      *dashboard.Dashboard
	mu             sync.RWMutex
	stats          *BuildStats
	pluginMgr      *plugin.PluginManager
	currentProcess *exec.Cmd
	processMu      sync.Mutex
	outputBinary   string
	projectDir     string
}

// BuildStats tracks rebuild statistics
type BuildStats struct {
	TotalRebuilds     int
	CacheHits         int
	CacheMisses       int
	ModuleRebuildTime map[string]time.Duration
	LastRebuildTime   time.Duration
	mu                sync.RWMutex
}

// NewOptimizer creates a new optimizer instance
func NewOptimizer(projectDir string) *Optimizer {
	// Initialize plugin manager
	pluginMgr := plugin.NewPluginManager()

	// Register available plugins
	pluginMgr.Register(plugin.NewGoPlugin(projectDir))
	pluginMgr.Register(plugin.NewWebpackPlugin("webpack.config.js"))
	pluginMgr.Register(plugin.NewVitePlugin("vite.config.js"))

	// Try to detect and activate a plugin
	if err := pluginMgr.DetectAndActivate(); err != nil {
		fmt.Printf("Warning: No build plugin detected: %v\n", err)
		fmt.Println("Hot reloader will run in analysis-only mode")
	} else {
		fmt.Printf("Detected build tool: %s\n", pluginMgr.GetActivePlugin().Name())
	}

	return &Optimizer{
		cache:        cache.NewModuleCache(),
		analyzer:     analyzer.NewDependencyAnalyzer(),
		depGraph:     analyzer.NewDependencyGraph(),
		dashboard:    dashboard.NewDashboard(),
		pluginMgr:    pluginMgr,
		outputBinary: "/tmp/hotreload_output",
		projectDir:   projectDir,
		stats: &BuildStats{
			ModuleRebuildTime: make(map[string]time.Duration),
		},
	}
}

// ProcessFileChange handles a file change event
func (o *Optimizer) ProcessFileChange(filePath string) error {
	o.mu.Lock()
	defer o.mu.Unlock()

	startTime := time.Now()

	// Check if file is in cache and still valid
	valid, err := o.cache.IsValid(filePath)
	if err != nil {
		return fmt.Errorf("error checking cache: %w", err)
	}

	if valid {
		o.stats.mu.Lock()
		o.stats.CacheHits++
		o.stats.mu.Unlock()
		o.dashboard.UpdateCacheHit(filePath)
		return nil
	}

	// Cache miss - need to rebuild
	o.stats.mu.Lock()
	o.stats.CacheMisses++
	o.stats.TotalRebuilds++
	o.stats.mu.Unlock()

	// Analyze dependencies
	deps, err := o.analyzer.AnalyzeDependencies(filePath)
	if err != nil {
		return fmt.Errorf("error analyzing dependencies: %w", err)
	}

	// Update dependency graph
	o.depGraph.AddDependency(filePath, deps)

	// Get all affected files (files that depend on this one)
	affectedFiles := o.depGraph.GetAllAffectedFiles(filePath)

	// Invalidate cache for affected files
	rebuildStart := time.Now()
	for _, file := range affectedFiles {
		o.cache.Invalidate(file)
	}

	// ACTUAL BUILD: Run the build plugin if available
	if o.pluginMgr.GetActivePlugin() != nil {
		fmt.Printf("\n🔨 Building (affected files: %d)...\n", len(affectedFiles))

		buildStart := time.Now()
		if err := o.pluginMgr.Build(affectedFiles); err != nil {
			fmt.Printf("❌ Build failed: %v\n", err)
			return fmt.Errorf("build failed: %w", err)
		}
		buildDuration := time.Since(buildStart)
		fmt.Printf("✅ Build successful (took %v)\n", buildDuration)

		// Only restart if using Go plugin (compiled binaries)
		if o.pluginMgr.GetActivePlugin().Name() == "go" {
			fmt.Println("🔄 Restarting application...")
			if err := o.restartProcess(); err != nil {
				fmt.Printf("⚠️  Failed to restart process: %v\n", err)
			} else {
				fmt.Println("✅ Application restarted successfully")
			}
		}
	}

	// Update cache for the changed file
	if err := o.cache.UpdateCache(filePath, deps); err != nil {
		return fmt.Errorf("error updating cache: %w", err)
	}

	duration := time.Since(rebuildStart)
	totalDuration := time.Since(startTime)

	o.stats.mu.Lock()
	o.stats.LastRebuildTime = totalDuration
	o.stats.mu.Unlock()

	// Update dashboard
	o.dashboard.UpdateRebuild(filePath, len(affectedFiles), duration)

	return nil
}

// GetStats returns current optimizer statistics
func (o *Optimizer) GetStats() *BuildStats {
	o.stats.mu.RLock()
	defer o.stats.mu.RUnlock()

	// Create a copy to avoid race conditions
	statsCopy := &BuildStats{
		TotalRebuilds:     o.stats.TotalRebuilds,
		CacheHits:         o.stats.CacheHits,
		CacheMisses:       o.stats.CacheMisses,
		LastRebuildTime:   o.stats.LastRebuildTime,
		ModuleRebuildTime: make(map[string]time.Duration),
	}

	for k, v := range o.stats.ModuleRebuildTime {
		statsCopy.ModuleRebuildTime[k] = v
	}

	return statsCopy
}

// GetDashboard returns the dashboard instance
func (o *Optimizer) GetDashboard() *dashboard.Dashboard {
	return o.dashboard
}

// PrintStats prints current statistics to stdout
func (o *Optimizer) PrintStats() {
	stats := o.GetStats()
	o.stats.mu.RLock()
	defer o.stats.mu.RUnlock()

	fmt.Println("\nHot Reload Optimizer Stats:")
	fmt.Printf("  Total Rebuilds: %d\n", stats.TotalRebuilds)
	fmt.Printf("  Cache Hits: %d\n", stats.CacheHits)
	fmt.Printf("  Cache Misses: %d\n", stats.CacheMisses)

	if stats.TotalRebuilds > 0 {
		hitRate := float64(stats.CacheHits) / float64(stats.CacheHits+stats.CacheMisses) * 100
		fmt.Printf("  Cache Hit Rate: %.2f%%\n", hitRate)
	}

	fmt.Printf("  Last Rebuild Time: %v\n", stats.LastRebuildTime)

	if len(stats.ModuleRebuildTime) > 0 {
		fmt.Println("\n  Module Rebuild Times:")
		for module, duration := range stats.ModuleRebuildTime {
			fmt.Printf("    %s: %v\n", module, duration)
		}
	}
}

// AnalyzeProject performs initial analysis of the entire project
func (o *Optimizer) AnalyzeProject(rootDir string) error {
	// This would recursively analyze all files in the project
	// and build the initial dependency graph
	return nil
}

// InitialBuild performs the first build and starts the application
func (o *Optimizer) InitialBuild() error {
	if o.pluginMgr.GetActivePlugin() == nil {
		fmt.Println("No build plugin available, skipping initial build")
		return nil
	}

	fmt.Println("\n🔨 Performing initial build...")
	buildStart := time.Now()

	// Build the project
	if err := o.pluginMgr.Build([]string{}); err != nil {
		fmt.Printf("❌ Initial build failed: %v\n", err)
		return fmt.Errorf("initial build failed: %w", err)
	}

	buildDuration := time.Since(buildStart)
	fmt.Printf("✅ Initial build successful (took %v)\n", buildDuration)

	// Start the process if it's a Go project
	if o.pluginMgr.GetActivePlugin().Name() == "go" {
		fmt.Println("▶️  Starting application...")
		if err := o.restartProcess(); err != nil {
			fmt.Printf("⚠️  Failed to start process: %v\n", err)
			return fmt.Errorf("failed to start process: %w", err)
		}
		fmt.Println("✅ Application started successfully\n")
	}

	return nil
}

// restartProcess stops the current process and starts a new one
func (o *Optimizer) restartProcess() error {
	o.processMu.Lock()
	defer o.processMu.Unlock()

	// Kill old process if it exists
	if o.currentProcess != nil && o.currentProcess.Process != nil {
		fmt.Printf("Stopping old process (PID: %d)...\n", o.currentProcess.Process.Pid)

		// Try graceful shutdown first
		if err := o.currentProcess.Process.Signal(os.Interrupt); err == nil {
			// Wait up to 2 seconds for graceful shutdown
			done := make(chan error)
			go func() {
				done <- o.currentProcess.Wait()
			}()

			select {
			case <-done:
				fmt.Println("Process stopped gracefully")
			case <-time.After(2 * time.Second):
				// Force kill if graceful shutdown times out
				fmt.Println("Graceful shutdown timed out, force killing...")
				o.currentProcess.Process.Kill()
				o.currentProcess.Wait()
			}
		} else {
			// If interrupt fails, just kill it
			o.currentProcess.Process.Kill()
			o.currentProcess.Wait()
		}
	}

	// Start new process
	cmd := exec.Command(o.outputBinary)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Dir = o.projectDir

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start process: %w", err)
	}

	o.currentProcess = cmd
	fmt.Printf("✅ Started new process with PID: %d\n", cmd.Process.Pid)

	return nil
}

// Shutdown gracefully stops the current running process
func (o *Optimizer) Shutdown() {
	o.processMu.Lock()
	defer o.processMu.Unlock()

	if o.currentProcess != nil && o.currentProcess.Process != nil {
		fmt.Printf("\nStopping process (PID: %d)...\n", o.currentProcess.Process.Pid)

		// Try graceful shutdown
		if err := o.currentProcess.Process.Signal(os.Interrupt); err == nil {
			done := make(chan error)
			go func() {
				done <- o.currentProcess.Wait()
			}()

			select {
			case <-done:
				fmt.Println("Process stopped gracefully")
			case <-time.After(2 * time.Second):
				fmt.Println("Force killing process...")
				o.currentProcess.Process.Kill()
				o.currentProcess.Wait()
			}
		} else {
			o.currentProcess.Process.Kill()
			o.currentProcess.Wait()
		}

		o.currentProcess = nil
	}
}
