package cmds

import (
	"fmt"
	"os"
	"sort"
	"strings"
)

type fileInfo struct {
	name        string
	size        int64
	permissions string
	fileType    string
	icon        string
	color       string
}

// formatSize converts size in bytes to human readable format
func formatSize(size int64) string {
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

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

	var files []fileInfo
	
	// Collect file information
	for _, entry := range entries {
		info, err := entry.Info()
		if err != nil {
			continue
		}

		var fileType, icon, color string
		if entry.IsDir() {
			fileType = "Directory"
			icon = iconFolder
			color = blue
		} else {
			mode := info.Mode()
			switch {
			case mode&os.ModeSymlink != 0:
				fileType = "Symlink"
				icon = iconSymlink
				color = yellow
			case mode&0111 != 0:
				fileType = "Executable"
				icon = iconExecutable
				color = cyan
			default:
				fileType = "File"
				icon = iconFile
				color = green
			}
		}

		files = append(files, fileInfo{
			name:        info.Name(),
			size:        info.Size(),
			permissions: info.Mode().String(),
			fileType:    fileType,
			icon:        icon,
			color:       color,
		})
	}

	// Sort files by type first, then by name
	sort.Slice(files, func(i, j int) bool {
		if files[i].fileType != files[j].fileType {
			return files[i].fileType < files[j].fileType
		}
		return files[i].name < files[j].name
	})

	// Find maximum lengths for column widths
	maxName := 4 // "NAME"
	maxSize := 4 // "SIZE"
	maxType := 4 // "TYPE"
	maxPerm := 11 // "PERMISSIONS"

	for _, f := range files {
		if len(f.name) > maxName {
			maxName = len(f.name)
		}
		sizeStr := formatSize(f.size)
		if len(sizeStr) > maxSize {
			maxSize = len(sizeStr)
		}
		if len(f.fileType) > maxType {
			maxType = len(f.fileType)
		}
	}

	// Print header
	fmt.Printf("┌%s┬%s┬%s┬%s┐\n",
		strings.Repeat("─", maxName+4),
		strings.Repeat("─", maxSize+2),
		strings.Repeat("─", maxType+2),
		strings.Repeat("─", maxPerm+2))
	
	fmt.Printf("│ %-*s │ %-*s │ %-*s │ %-*s │\n",
		maxName+2, "NAME",
		maxSize, "SIZE",
		maxType, "TYPE",
		maxPerm, "PERMISSIONS")
	
	fmt.Printf("├%s┼%s┼%s┼%s┤\n",
		strings.Repeat("─", maxName+4),
		strings.Repeat("─", maxSize+2),
		strings.Repeat("─", maxType+2),
		strings.Repeat("─", maxPerm+2))

	// Print files
	for _, f := range files {
		sizeStr := formatSize(f.size)
		fmt.Printf("│ %s%s %s%-*s%s │ %*s │ %-*s │ %-*s │\n",
			f.color, f.icon, reset,
			maxName, f.name,
			reset,
			maxSize, sizeStr,
			maxType, f.fileType,
			maxPerm, f.permissions)
	}

	// Print footer
	fmt.Printf("└%s┴%s┴%s┴%s┘\n",
		strings.Repeat("─", maxName+4),
		strings.Repeat("─", maxSize+2),
		strings.Repeat("─", maxType+2),
		strings.Repeat("─", maxPerm+2))
}
