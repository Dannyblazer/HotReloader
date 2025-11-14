package dashboard

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// Dashboard displays real-time rebuild metrics
type Dashboard struct {
	mu              sync.RWMutex
	events          []Event
	maxEvents       int
	lastUpdate      time.Time
	totalCacheHits  int
	totalRebuilds   int
	totalAffected   int
}

// Event represents a rebuild event
type Event struct {
	Timestamp     time.Time
	FilePath      string
	AffectedCount int
	Duration      time.Duration
	EventType     EventType
}

// EventType defines the type of event
type EventType int

const (
	RebuildEvent EventType = iota
	CacheHitEvent
)

// NewDashboard creates a new dashboard instance
func NewDashboard() *Dashboard {
	return &Dashboard{
		events:    make([]Event, 0),
		maxEvents: 50, // Keep last 50 events
	}
}

// UpdateRebuild records a rebuild event
func (d *Dashboard) UpdateRebuild(filePath string, affectedCount int, duration time.Duration) {
	d.mu.Lock()
	defer d.mu.Unlock()

	event := Event{
		Timestamp:     time.Now(),
		FilePath:      filePath,
		AffectedCount: affectedCount,
		Duration:      duration,
		EventType:     RebuildEvent,
	}

	d.events = append(d.events, event)
	d.totalRebuilds++
	d.totalAffected += affectedCount
	d.lastUpdate = time.Now()

	// Keep only the last N events
	if len(d.events) > d.maxEvents {
		d.events = d.events[len(d.events)-d.maxEvents:]
	}

	d.displayEvent(event)
}

// UpdateCacheHit records a cache hit event
func (d *Dashboard) UpdateCacheHit(filePath string) {
	d.mu.Lock()
	defer d.mu.Unlock()

	event := Event{
		Timestamp: time.Now(),
		FilePath:  filePath,
		EventType: CacheHitEvent,
	}

	d.events = append(d.events, event)
	d.totalCacheHits++
	d.lastUpdate = time.Now()

	// Keep only the last N events
	if len(d.events) > d.maxEvents {
		d.events = d.events[len(d.events)-d.maxEvents:]
	}

	d.displayEvent(event)
}

// displayEvent prints an event to the console
func (d *Dashboard) displayEvent(event Event) {
	timestamp := event.Timestamp.Format("15:04:05")

	switch event.EventType {
	case RebuildEvent:
		fmt.Printf("[%s] REBUILD: %s (affected: %d files, took: %v)\n",
			timestamp, event.FilePath, event.AffectedCount, event.Duration)
	case CacheHitEvent:
		fmt.Printf("[%s] CACHE HIT: %s (skipped rebuild)\n",
			timestamp, event.FilePath)
	}
}

// PrintSummary displays a summary of all events
func (d *Dashboard) PrintSummary() {
	d.mu.RLock()
	defer d.mu.RUnlock()

	separator := strings.Repeat("=", 60)
	fmt.Println("\n" + separator)
	fmt.Println("HOT RELOAD OPTIMIZER - DASHBOARD")
	fmt.Println(separator)

	if d.totalRebuilds == 0 && d.totalCacheHits == 0 {
		fmt.Println("No events yet. Waiting for file changes...")
		return
	}

	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Total Rebuilds:  %d\n", d.totalRebuilds)
	fmt.Printf("  Cache Hits:      %d\n", d.totalCacheHits)
	fmt.Printf("  Total Affected:  %d files\n", d.totalAffected)

	if d.totalRebuilds > 0 {
		avgAffected := float64(d.totalAffected) / float64(d.totalRebuilds)
		fmt.Printf("  Avg Affected:    %.2f files per rebuild\n", avgAffected)
	}

	total := d.totalRebuilds + d.totalCacheHits
	if total > 0 {
		cacheHitRate := float64(d.totalCacheHits) / float64(total) * 100
		fmt.Printf("  Cache Hit Rate:  %.2f%%\n", cacheHitRate)
	}

	fmt.Printf("\nRecent Events (last %d):\n", min(len(d.events), 10))
	recentEvents := d.events
	if len(recentEvents) > 10 {
		recentEvents = recentEvents[len(recentEvents)-10:]
	}

	for i := len(recentEvents) - 1; i >= 0; i-- {
		event := recentEvents[i]
		timestamp := event.Timestamp.Format("15:04:05")

		switch event.EventType {
		case RebuildEvent:
			fmt.Printf("  [%s] REBUILD: %s (%d files, %v)\n",
				timestamp, event.FilePath, event.AffectedCount, event.Duration)
		case CacheHitEvent:
			fmt.Printf("  [%s] CACHE HIT: %s (cached)\n",
				timestamp, event.FilePath)
		}
	}

	fmt.Println(separator + "\n")
}

// GetMetrics returns current dashboard metrics
func (d *Dashboard) GetMetrics() map[string]interface{} {
	d.mu.RLock()
	defer d.mu.RUnlock()

	return map[string]interface{}{
		"total_rebuilds":   d.totalRebuilds,
		"total_cache_hits": d.totalCacheHits,
		"total_affected":   d.totalAffected,
		"last_update":      d.lastUpdate,
		"event_count":      len(d.events),
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
