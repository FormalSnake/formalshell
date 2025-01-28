package main

import (
	"fmt"
	"os"
	"sort"
)

// customLS is a replacement for the `ls` command that shows files and folders with colors and icons.
func customLS() {
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
