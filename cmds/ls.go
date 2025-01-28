package cmds

import (
	"fmt"
	"os"
	"sort"
)

// ANSI color codes
const (
	reset   = "\033[0m"
	blue    = "\033[34m" // Directories
	cyan    = "\033[36m" // Executable files
	green   = "\033[32m" // Regular files
	yellow  = "\033[33m" // Symlinks
	magenta = "\033[35m" // Special files
)

// Nerd Fonts icons
const (
	iconFolder     = "" // Nerd Font icon for folders
	iconFile       = "󰈙" // Nerd Font icon for files
	iconExecutable = "" // Nerd Font icon for executables
	iconSymlink    = "" // Nerd Font icon for symlinks
)

// customLS is a replacement for the `ls` command that shows files and folders with colors and icons.
func CustomLS() {
	// Get current directory
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(wd)
	if err != nil {
		fmt.Println("Error:", err)
		return
	}

	// Sort entries alphabetically
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].Name() < entries[j].Name()
	})

	// Display each entry with appropriate color and icon
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() {
			// Directory: blue color with folder icon
			fmt.Printf("%s%s %s%s\n", blue, iconFolder, name, reset)
		} else {
			info, err := entry.Info()
			if err != nil {
				fmt.Println("Error:", err)
				continue
			}

			mode := info.Mode()
			switch {
			case mode&os.ModeSymlink != 0:
				// Symlink: yellow color with symlink icon
				fmt.Printf("%s%s %s%s\n", yellow, iconSymlink, name, reset)
			case mode&0111 != 0: // Executable files
				// Executable: cyan color with executable icon
				fmt.Printf("%s%s %s%s\n", cyan, iconExecutable, name, reset)
			default:
				// Regular files: green color with file icon
				fmt.Printf("%s%s %s%s\n", green, iconFile, name, reset)
			}
		}
	}
}
