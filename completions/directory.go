package completions

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

// CreateCompleter returns a readline.PrefixCompleter for the shell
func CreateCompleter(commandHistory map[string]bool) *readline.PrefixCompleter {
	var completions []readline.PrefixCompleterInterface
	
	// Add built-in commands
	cdCompleter := readline.PcItem("cd",
		readline.PcItemDynamic(func(line string) []string {
			// Get current directory entries
			entries, err := os.ReadDir(".")
			if err != nil {
				return nil
			}

			// Filter for directories only
			var dirs []string
			for _, entry := range entries {
				if entry.IsDir() {
					dirs = append(dirs, entry.Name())
				}
			}
			return dirs
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
