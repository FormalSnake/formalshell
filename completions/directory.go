package completions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

// getDirCompletions returns directory completions for a given path prefix
func getDirCompletions(prefix string) []string {
	// If prefix is empty, use current directory
	searchDir := "."
	searchPrefix := ""

	// Handle home directory expansion
	if strings.HasPrefix(prefix, "~/") {
		if homeDir, err := os.UserHomeDir(); err == nil {
			prefix = filepath.Join(homeDir, prefix[2:])
		}
	}

	// Handle special paths
	if prefix != "" {
		switch {
		case strings.HasPrefix(prefix, "~/"):
			// Home directory
			if homeDir, err := os.UserHomeDir(); err == nil {
				searchDir = filepath.Join(homeDir, filepath.Dir(prefix[2:]))
				searchPrefix = filepath.Base(prefix)
			}
		case strings.HasPrefix(prefix, "../"):
			// Parent directory
			if currentDir, err := os.Getwd(); err == nil {
				searchDir = filepath.Join(currentDir, filepath.Dir(prefix))
				searchPrefix = filepath.Base(prefix)
			}
		case strings.HasPrefix(prefix, "./"):
			// Current directory
			if currentDir, err := os.Getwd(); err == nil {
				searchDir = filepath.Join(currentDir, filepath.Dir(prefix[2:]))
				searchPrefix = filepath.Base(prefix)
			}
		default:
			// Regular path
			searchDir = filepath.Dir(prefix)
			if searchDir == "." {
				searchPrefix = prefix
			} else {
				searchPrefix = filepath.Base(prefix)
			}
		}
	}

	// Read the directory
	entries, err := os.ReadDir(searchDir)
	if err != nil {
		return nil
	}

	var completions []string
	for _, entry := range entries {
		if entry.IsDir() {
			name := entry.Name()
			if searchPrefix == "" || strings.HasPrefix(strings.ToLower(name), strings.ToLower(searchPrefix)) {
				if searchDir == "." {
					completions = append(completions, name)
				} else {
					completions = append(completions, filepath.Join(filepath.Dir(prefix), name))
				}
			}
		}
	}

	return completions
}

// CreateCompleter returns a readline.PrefixCompleter for the shell
func CreateCompleter(commandHistory map[string]bool) *readline.PrefixCompleter {
	var completions []readline.PrefixCompleterInterface

	// Add built-in commands with directory completion for cd
	cdCompleter := readline.PcItem("cd",
		readline.PcItemDynamic(func(line string) []string {
			parts := strings.Fields(line)
			if len(parts) <= 1 {
				// Show all directories in current directory
				return getDirCompletions("")
			}
			// Get the partial path after "cd "
			partial := strings.TrimPrefix(line, "cd ")
			partial = strings.TrimSpace(partial)
			return getDirCompletions(partial)
		}))

	completions = append(completions, cdCompleter)

	// Add command history completions
	for cmd := range commandHistory {
		if !strings.HasPrefix(cmd, "cd ") { // Skip cd commands from history
			completions = append(completions, readline.PcItem(cmd))
		}
	}

	return readline.NewPrefixCompleter(completions...)
}
