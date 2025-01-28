package main

import (
	"fmt"
	"os"
	"path/filepath"
)

// handleCD implements the 'cd' command to change directories.
func handleCD(args []string) {
	if len(args) < 1 {
		fmt.Println("cd: missing argument")
		return
	}

	// Resolve relative paths
	path, err := filepath.Abs(args[0])
	if err != nil {
		fmt.Println("cd:", err)
		return
	}

	// Change directory
	if err := os.Chdir(path); err != nil {
		fmt.Println("cd:", err)
	}
}
