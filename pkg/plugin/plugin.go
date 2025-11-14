package plugin

import (
	"fmt"
	"os/exec"
	"time"
)

// BuildPlugin defines the interface for build tool plugins
type BuildPlugin interface {
	Name() string
	Detect() bool
	Build(files []string) error
	GetBuildTime() time.Duration
}

// WebpackPlugin implements Webpack integration
type WebpackPlugin struct {
	configPath string
	lastBuildTime time.Duration
}

// NewWebpackPlugin creates a new Webpack plugin
func NewWebpackPlugin(configPath string) *WebpackPlugin {
	return &WebpackPlugin{
		configPath: configPath,
	}
}

// Name returns the plugin name
func (w *WebpackPlugin) Name() string {
	return "webpack"
}

// Detect checks if Webpack is available
func (w *WebpackPlugin) Detect() bool {
	_, err := exec.LookPath("webpack")
	return err == nil
}

// Build runs webpack build for specified files
func (w *WebpackPlugin) Build(files []string) error {
	start := time.Now()

	cmd := exec.Command("webpack", "--config", w.configPath)
	output, err := cmd.CombinedOutput()

	w.lastBuildTime = time.Since(start)

	if err != nil {
		return fmt.Errorf("webpack build failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetBuildTime returns the last build time
func (w *WebpackPlugin) GetBuildTime() time.Duration {
	return w.lastBuildTime
}

// VitePlugin implements Vite integration
type VitePlugin struct {
	configPath string
	lastBuildTime time.Duration
}

// NewVitePlugin creates a new Vite plugin
func NewVitePlugin(configPath string) *VitePlugin {
	return &VitePlugin{
		configPath: configPath,
	}
}

// Name returns the plugin name
func (v *VitePlugin) Name() string {
	return "vite"
}

// Detect checks if Vite is available
func (v *VitePlugin) Detect() bool {
	_, err := exec.LookPath("vite")
	return err == nil
}

// Build runs vite build for specified files
func (v *VitePlugin) Build(files []string) error {
	start := time.Now()

	cmd := exec.Command("vite", "build")
	output, err := cmd.CombinedOutput()

	v.lastBuildTime = time.Since(start)

	if err != nil {
		return fmt.Errorf("vite build failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetBuildTime returns the last build time
func (v *VitePlugin) GetBuildTime() time.Duration {
	return v.lastBuildTime
}

// GoPlugin implements Go build integration
type GoPlugin struct {
	modulePath string
	lastBuildTime time.Duration
}

// NewGoPlugin creates a new Go plugin
func NewGoPlugin(modulePath string) *GoPlugin {
	return &GoPlugin{
		modulePath: modulePath,
	}
}

// Name returns the plugin name
func (g *GoPlugin) Name() string {
	return "go"
}

// Detect checks if Go is available
func (g *GoPlugin) Detect() bool {
	_, err := exec.LookPath("go")
	return err == nil
}

// Build runs go build for specified files
func (g *GoPlugin) Build(files []string) error {
	start := time.Now()

	cmd := exec.Command("go", "build", "-o", "/tmp/output", g.modulePath)
	output, err := cmd.CombinedOutput()

	g.lastBuildTime = time.Since(start)

	if err != nil {
		return fmt.Errorf("go build failed: %w\nOutput: %s", err, string(output))
	}

	return nil
}

// GetBuildTime returns the last build time
func (g *GoPlugin) GetBuildTime() time.Duration {
	return g.lastBuildTime
}

// PluginManager manages build plugins
type PluginManager struct {
	plugins []BuildPlugin
	active  BuildPlugin
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make([]BuildPlugin, 0),
	}
}

// Register adds a plugin to the manager
func (pm *PluginManager) Register(plugin BuildPlugin) {
	pm.plugins = append(pm.plugins, plugin)
}

// DetectAndActivate detects and activates an available plugin
func (pm *PluginManager) DetectAndActivate() error {
	for _, plugin := range pm.plugins {
		if plugin.Detect() {
			pm.active = plugin
			return nil
		}
	}
	return fmt.Errorf("no suitable build plugin found")
}

// GetActivePlugin returns the currently active plugin
func (pm *PluginManager) GetActivePlugin() BuildPlugin {
	return pm.active
}

// Build runs the active plugin's build
func (pm *PluginManager) Build(files []string) error {
	if pm.active == nil {
		return fmt.Errorf("no active plugin")
	}
	return pm.active.Build(files)
}
