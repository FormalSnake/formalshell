package completions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

// GetDirCompletions returns a list of subdirectories for the given path
func GetDirCompletions(path string) []string {
	var completions []string
	
	// Handle home directory in completion
	if strings.HasPrefix(path, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			path = filepath.Join(homeDir, path[2:])
		}
	}
	
	// If path is empty or ".", use current directory
	if path == "" || path == "." {
		path = "."
	}
	
	// Get the directory to search in and the prefix to match
	var searchDir string
	var prefix string
	
	if filepath.IsAbs(path) {
		searchDir = filepath.Dir(path)
		prefix = filepath.Base(path)
	} else {
		// For relative paths, search in current directory
		currentDir, _ := os.Getwd()
		searchDir = currentDir
		prefix = path
	}
	
	// List all directories
	if entries, err := os.ReadDir(searchDir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() && strings.HasPrefix(strings.ToLower(name), strings.ToLower(prefix)) {
				if len(prefix) > 0 {
					completions = append(completions, name[len(prefix):])
				} else {
					completions = append(completions, name)
				}
			}
		}
	}
	
	return completions
}

// CreateCompleter returns a readline.PrefixCompleter for the shell
func CreateCompleter(commandHistory map[string]bool) *readline.PrefixCompleter {
	var completions []readline.PrefixCompleterInterface
	
	// Add command history completions
	for cmd := range commandHistory {
		completions = append(completions, readline.PcItem(cmd))
	}
	
	// Add cd command with directory completion
	cdCompleter := readline.PcItem("cd",
		readline.PcItemDynamic(func(path string) []string {
			return GetDirCompletions(path)
		}))
	
	completions = append(completions, cdCompleter)
	
	return readline.NewPrefixCompleter(completions...)
}
