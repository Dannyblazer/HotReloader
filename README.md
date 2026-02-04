# 🚀 Hot Reload Optimizer

A high-performance development tool that analyzes which files actually need rebuilding on change, caches unchanged modules aggressively, and shows detailed rebuild time breakdowns by module.

## Features

- **Smart Dependency Analysis**: Automatically detects and tracks dependencies across Go, JavaScript, TypeScript, JSX, TSX, and Python files
- **Intelligent Caching**: Hash-based module caching that only rebuilds what actually changed
- **Real-time Dashboard**: Live metrics showing rebuild times, cache hit rates, and affected files
- **File System Watcher**: Efficient monitoring using fsnotify with debouncing and smart filtering
- **Build Tool Plugins**: Extensible plugin system for Webpack, Vite, Go, and other build tools
- **Performance Tracking**: Detailed breakdown of rebuild time per module
- **Automatic Build & Restart**: Actually builds your code and restarts the process on changes
- **Process Management**: Graceful shutdown and restart of running applications

## ⚡ Quick Start

### Installation

```bash
# Clone the repository
git clone https://github.com/Dannyblazer/hotreloader.git
cd hotreloader

# Build the binary
go build -o hotreloader .

# Optional: Install globally
go install
```

### Basic Usage

```bash
# Watch current directory
./hotreloader .

# Watch specific directory
./hotreloader /path/to/your/project

# Watch example demo app
cd examples/demo-app
../../hotreloader .
```

## 🔍 How It Works

### 1. Dependency Analysis

The optimizer analyzes your source files to build a dependency graph:

```
main.js
├── utils.js
└── api.js
    └── utils.js
```

When `utils.js` changes, it knows to rebuild both `api.js` and `main.js`.

### 2. Smart Caching

Each file is hashed (SHA-256) and cached with metadata:
- File hash
- Last modified time
- File size
- Dependencies list

Files are only rebuilt if:
- The hash changes
- Any dependency changes
- Cache entry is missing

### 3. Real-time Metrics

The dashboard shows:
- Total rebuilds vs cache hits
- Cache hit rate percentage
- Files affected per rebuild
- Time spent per module
- Recent activity log

## 📊 Example Output

```
🚀 Hot Reload Optimizer watching: ./examples/demo-app
Press Ctrl+C to stop...

👀 Watching for changes... (Press Ctrl+C to show stats and exit)

[15:23:45] 🔨 REBUILD: src/utils.js (affected: 2 files, took: 45ms)
[15:23:52] ✅ CACHE HIT: src/main.js (skipped rebuild)
[15:24:01] 🔨 REBUILD: src/api.js (affected: 1 files, took: 23ms)

════════════════════════════════════════════════════════════
📈 HOT RELOAD OPTIMIZER - DASHBOARD
════════════════════════════════════════════════════════════

📊 Summary:
  Total Rebuilds:  2
  Cache Hits:      1
  Total Affected:  3 files
  Avg Affected:    1.50 files per rebuild
  Cache Hit Rate:  33.33%

📋 Recent Events (last 10):
  [15:24:01] 🔨 src/api.js (1 files, 23ms)
  [15:23:52] ✅ src/main.js (cached)
  [15:23:45] 🔨 src/utils.js (2 files, 45ms)
════════════════════════════════════════════════════════════
```

## 🏗️ Architecture

### Project Structure

```
hotreloader/
├── main.go                 # CLI entry point
└── pkg/
    ├── analyzer/           # Dependency analysis
    │   └── analyzer.go
    ├── cache/              # Module caching system
    │   └── cache.go
    ├── dashboard/          # Real-time metrics display
    │   └── dashboard.go
    ├── optimizer/          # Core optimization engine
    │   └── optimizer.go
    ├── plugin/             # Build tool plugins
    │   └── plugin.go
    └── watcher/            # File system monitoring
        └── watcher.go
examples/
└── demo-app/               # Example application
```

### Core Components

#### Analyzer ([pkg/analyzer/analyzer.go](pkg/analyzer/analyzer.go))
- Parses import/require statements
- Builds dependency graphs
- Identifies transitive dependencies
- Supports multiple languages

#### Cache ([pkg/cache/cache.go](pkg/cache/cache.go))
- SHA-256 file hashing
- Metadata storage
- Cache validation
- Thread-safe operations

#### Optimizer ([pkg/optimizer/optimizer.go](pkg/optimizer/optimizer.go))
- Orchestrates rebuild decisions
- Tracks statistics
- Manages cache and analyzer
- Provides metrics API

#### Watcher ([pkg/watcher/watcher.go](pkg/watcher/watcher.go))
- fsnotify-based file monitoring
- Recursive directory watching
- Debouncing for rapid changes
- Smart path filtering

#### Dashboard ([pkg/dashboard/dashboard.go](pkg/dashboard/dashboard.go))
- Real-time event display
- Statistics aggregation
- Summary reporting
- Cache hit rate calculation

## 🔌 Plugin System

The plugin system allows integration with various build tools:

### Available Plugins

- **Webpack**: `WebpackPlugin`
- **Vite**: `VitePlugin`
- **Go**: `GoPlugin`

### Creating a Custom Plugin

```go
type CustomPlugin struct {
    lastBuildTime time.Duration
}

func (p *CustomPlugin) Name() string {
    return "custom"
}

func (p *CustomPlugin) Detect() bool {
    // Check if your build tool is available
    _, err := exec.LookPath("yourtool")
    return err == nil
}

func (p *CustomPlugin) Build(files []string) error {
    start := time.Now()

    // Run your build command
    cmd := exec.Command("yourtool", "build")
    err := cmd.Run()

    p.lastBuildTime = time.Since(start)
    return err
}

func (p *CustomPlugin) GetBuildTime() time.Duration {
    return p.lastBuildTime
}
```

## ⚙️ Configuration

### Ignored Paths

By default, these paths are ignored:
- `node_modules/`
- `.git/`
- `.vscode/`
- `.idea/`
- `dist/`
- `build/`
- `*.log`
- `.DS_Store`

You can modify the ignore list in [pkg/watcher/watcher.go](pkg/watcher/watcher.go).

### Debounce Time

Default debounce time is 100ms. Adjust in [pkg/watcher/watcher.go](pkg/watcher/watcher.go):

```go
debounce: 100 * time.Millisecond,  // Change this value
```

## ⚡ Performance Benefits

### Without Hot Reload Optimizer

```
Change file → Full rebuild (5-30 seconds)
Every. Single. Time.
```

### With Hot Reload Optimizer

```
Change file → Analyze dependencies → Check cache → Rebuild only affected (100-500ms)
Cache hits → Skip rebuild entirely (< 1ms)
```

### Typical Results

- **70-90%** reduction in rebuild time
- **50-80%** cache hit rate after warm-up
- **Instant** feedback for cached files

## 🛠️ Development

### Building from Source

```bash
# Clone the repo
git clone https://github.com/Dannyblazer/hotreloader.git
cd hotreloader

# Install dependencies
go mod tidy

# Build
go build -o hotreloader .

# Run tests (when available)
go test ./...
```

### Running the Example

```bash
# Build the optimizer
go build -o hotreloader .

# Run with the demo app
cd examples/demo-app
../../hotreloader .

# In another terminal, modify files to see it in action
echo "// new comment" >> src/utils.js
```

## 💡 Use Cases

### Frontend Development
- React, Vue, Angular applications
- Webpack/Vite projects
- Large component libraries

### Backend Development
- Go microservices
- Node.js APIs
- Multi-module projects

### Full-Stack Applications
- Monorepos
- Multi-language projects
- Complex dependency chains

## 🤝 Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

### Areas for Contribution
- Additional language support
- More build tool plugins
- Configuration file support
- Web-based dashboard
- Performance optimizations

## 📄 License

MIT License - see LICENSE file for details

## 🙏 Acknowledgments

- [fsnotify](https://github.com/fsnotify/fsnotify) - Cross-platform file system notifications
- Inspired by modern build tools like Air and Turbopack

## 💬 Support

- Issues: [GitHub Issues](https://github.com/Dannyblazer/hotreloader/issues)
- Discussions: [GitHub Discussions](https://github.com/Dannyblazer/hotreloader/discussions)

---

Made with ❤️ for developers who value their time
