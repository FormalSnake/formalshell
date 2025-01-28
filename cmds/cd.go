package cmds

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var dirDB = NewDirectoryDB()

// HandleCD implements the 'cd' command to change directories.
func HandleCD(args []string) {
	if len(args) < 1 {
		// Change to home directory if no args
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("cd:", err)
			return
		}
		if err := os.Chdir(homeDir); err != nil {
			fmt.Println("cd:", err)
		}
		dirDB.AddVisit(homeDir)
		return
	}

	path := args[0]

	// Handle home directory expansion
	if strings.HasPrefix(path, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			fmt.Println("cd:", err)
			return
		}
		path = filepath.Join(homeDir, path[2:])
	}

	// Try smart directory matching if path doesn't exist
	if _, err := os.Stat(path); os.IsNotExist(err) {
		smartPath := dirDB.FindMatch(path)
		if smartPath != path {
			path = smartPath
		}
	}

	// Resolve relative paths
	path, err := filepath.Abs(path)
	if err != nil {
		fmt.Println("cd:", err)
		return
	}

	// Change directory
	if err := os.Chdir(path); err != nil {
		fmt.Println("cd:", err)
		return
	}

	// Record successful directory change
	dirDB.AddVisit(path)
}
