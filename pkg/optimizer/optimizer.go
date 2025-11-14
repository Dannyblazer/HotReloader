package optimizer

import (
	"fmt"
	"sync"
	"time"

	"hotreloader/pkg/analyzer"
	"hotreloader/pkg/cache"
	"hotreloader/pkg/dashboard"
)

// Optimizer is the core hot reload optimizer
type Optimizer struct {
	cache     *cache.ModuleCache
	analyzer  *analyzer.DependencyAnalyzer
	depGraph  *analyzer.DependencyGraph
	dashboard *dashboard.Dashboard
	mu        sync.RWMutex
	stats     *BuildStats
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
func NewOptimizer() *Optimizer {
	return &Optimizer{
		cache:     cache.NewModuleCache(),
		analyzer:  analyzer.NewDependencyAnalyzer(),
		depGraph:  analyzer.NewDependencyGraph(),
		dashboard: dashboard.NewDashboard(),
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

	// Simulate rebuild and track time per module
	rebuildStart := time.Now()
	for _, file := range affectedFiles {
		moduleStart := time.Now()

		// This is where actual rebuild would happen
		// For now, we'll just invalidate the cache
		o.cache.Invalidate(file)

		moduleDuration := time.Since(moduleStart)
		o.stats.mu.Lock()
		o.stats.ModuleRebuildTime[file] = moduleDuration
		o.stats.mu.Unlock()
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
