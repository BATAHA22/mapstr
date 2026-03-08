package git

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/mapstr/mapstr/internal/parser"
)

const cacheFile = ".mapstr-cache.json"

// ChangedFiles returns the list of files changed since the last commit.
// If not in a git repo, it returns an error.
func ChangedFiles(root string) ([]string, error) {
	cmd := exec.Command("git", "diff", "HEAD~1", "--name-only")
	cmd.Dir = root

	out, err := cmd.Output()
	if err != nil {
		// Try against empty tree (first commit or no commits)
		cmd = exec.Command("git", "diff", "--cached", "--name-only")
		cmd.Dir = root
		out, err = cmd.Output()
		if err != nil {
			return nil, fmt.Errorf("git: diff: %w", err)
		}
	}

	var files []string
	for _, line := range strings.Split(strings.TrimSpace(string(out)), "\n") {
		line = strings.TrimSpace(line)
		if line != "" {
			files = append(files, line)
		}
	}

	return files, nil
}

// IsGitRepo checks if the given directory is inside a git repository.
func IsGitRepo(root string) bool {
	cmd := exec.Command("git", "rev-parse", "--is-inside-work-tree")
	cmd.Dir = root
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "true"
}

// SaveCache writes parsed file nodes to the cache file.
func SaveCache(root string, nodes []*parser.FileNode) error {
	path := filepath.Join(root, cacheFile)
	data, err := json.MarshalIndent(nodes, "", "  ")
	if err != nil {
		return fmt.Errorf("git: cache marshal: %w", err)
	}
	return os.WriteFile(path, data, 0644)
}

// LoadCache reads previously cached file nodes.
func LoadCache(root string) ([]*parser.FileNode, error) {
	path := filepath.Join(root, cacheFile)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("git: cache read: %w", err)
	}

	var nodes []*parser.FileNode
	if err := json.Unmarshal(data, &nodes); err != nil {
		return nil, fmt.Errorf("git: cache unmarshal: %w", err)
	}

	return nodes, nil
}

// MergeNodes merges newly parsed nodes into the cached set,
// replacing any cached nodes whose paths match a new node.
func MergeNodes(cached, fresh []*parser.FileNode) []*parser.FileNode {
	freshIndex := map[string]*parser.FileNode{}
	for _, n := range fresh {
		freshIndex[n.Path] = n
	}

	var result []*parser.FileNode
	seen := map[string]bool{}

	// Add all fresh nodes first
	for _, n := range fresh {
		result = append(result, n)
		seen[n.Path] = true
	}

	// Add cached nodes that weren't replaced
	for _, n := range cached {
		if !seen[n.Path] {
			result = append(result, n)
		}
	}

	return result
}
