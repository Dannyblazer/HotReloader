package analyzer

import (
	"bufio"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// DependencyAnalyzer analyzes file dependencies
type DependencyAnalyzer struct {
	importPatterns map[string]*regexp.Regexp
}

// NewDependencyAnalyzer creates a new dependency analyzer
func NewDependencyAnalyzer() *DependencyAnalyzer {
	return &DependencyAnalyzer{
		importPatterns: map[string]*regexp.Regexp{
			".go":   regexp.MustCompile(`^\s*import\s+(?:"([^"]+)"|([a-zA-Z_]\w*)\s+"([^"]+)")`),
			".js":   regexp.MustCompile(`(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\))`),
			".ts":   regexp.MustCompile(`(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\))`),
			".jsx":  regexp.MustCompile(`(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\))`),
			".tsx":  regexp.MustCompile(`(?:import\s+.*?\s+from\s+['"]([^'"]+)['"]|require\s*\(\s*['"]([^'"]+)['"]\s*\))`),
			".py":   regexp.MustCompile(`^\s*(?:from\s+(\S+)\s+import|import\s+(\S+))`),
		},
	}
}

// AnalyzeDependencies extracts dependencies from a file
func (a *DependencyAnalyzer) AnalyzeDependencies(filePath string) ([]string, error) {
	ext := filepath.Ext(filePath)
	pattern, ok := a.importPatterns[ext]
	if !ok {
		// Unsupported file type, no dependencies
		return []string{}, nil
	}

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	dependencies := make(map[string]bool)
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := scanner.Text()
		matches := pattern.FindAllStringSubmatch(line, -1)

		for _, match := range matches {
			for i := 1; i < len(match); i++ {
				if match[i] != "" {
					dep := a.normalizeDependency(match[i], ext)
					if dep != "" {
						dependencies[dep] = true
					}
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	// Convert map to slice
	deps := make([]string, 0, len(dependencies))
	for dep := range dependencies {
		deps = append(deps, dep)
	}

	return deps, nil
}

// normalizeDependency normalizes a dependency path
func (a *DependencyAnalyzer) normalizeDependency(dep, ext string) string {
	dep = strings.TrimSpace(dep)

	// Skip empty dependencies
	if dep == "" {
		return ""
	}

	// For relative imports in JS/TS, keep them as-is
	if ext == ".js" || ext == ".ts" || ext == ".jsx" || ext == ".tsx" {
		return dep
	}

	return dep
}

// DependencyGraph represents a graph of file dependencies
type DependencyGraph struct {
	graph map[string][]string
}

// NewDependencyGraph creates a new dependency graph
func NewDependencyGraph() *DependencyGraph {
	return &DependencyGraph{
		graph: make(map[string][]string),
	}
}

// AddDependency adds a dependency edge to the graph
func (g *DependencyGraph) AddDependency(file string, deps []string) {
	g.graph[file] = deps
}

// GetDependents returns all files that depend on the given file
func (g *DependencyGraph) GetDependents(file string) []string {
	dependents := []string{}
	for f, deps := range g.graph {
		for _, dep := range deps {
			if strings.Contains(dep, filepath.Base(file)) || dep == file {
				dependents = append(dependents, f)
				break
			}
		}
	}
	return dependents
}

// GetAllAffectedFiles returns all files affected by a change (including transitive deps)
func (g *DependencyGraph) GetAllAffectedFiles(file string) []string {
	visited := make(map[string]bool)
	affected := []string{}

	var traverse func(string)
	traverse = func(f string) {
		if visited[f] {
			return
		}
		visited[f] = true
		affected = append(affected, f)

		for _, dependent := range g.GetDependents(f) {
			traverse(dependent)
		}
	}

	traverse(file)
	return affected
}
