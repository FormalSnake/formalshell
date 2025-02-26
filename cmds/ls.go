package cmds

import (
	"fmt"
	"os"
	"path/filepath"
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
	gray    = "\033[38;5;242m" // Table borders
)

// Nerd Fonts icons
const (
	iconFolder     = "" // Nerd Font icon for folders
	iconFile       = "󰈙" // Nerd Font icon for files
	iconExecutable = "" // Nerd Font icon for executables
	iconSymlink    = "" // Nerd Font icon for symlinks
)

// customLS is a replacement for the `ls` command that shows files and folders with colors and icons.
func CustomLS(args ...string) {
	// Determine target directory
	targetDir := "."
	if len(args) > 0 {
		targetDir = args[0]
	}

	// Handle ~ in path
	if strings.HasPrefix(targetDir, "~") {
		homeDir, err := os.UserHomeDir()
		if err == nil {
			targetDir = strings.Replace(targetDir, "~", homeDir, 1)
		}
	}

	// Get absolute path
	targetDir, err := filepath.Abs(targetDir)
	if err != nil {
		fmt.Printf("Error resolving path: %v\n", err)
		return
	}

	// Read directory contents
	entries, err := os.ReadDir(targetDir)
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

		var fileType, color string
		mode := info.Mode()
		isSymlink := mode&os.ModeSymlink != 0
		isExecutable := mode&0111 != 0
		isDir := entry.IsDir()

		// Get appropriate icon
		icon := GetFileIcon(info.Name(), isDir, isExecutable, isSymlink)

		// Set type and color
		switch {
		case isDir:
			fileType = "Directory"
			color = blue
		case isSymlink:
			fileType = "Symlink"
			color = yellow
		case isExecutable:
			fileType = "Executable"
			color = cyan
		default:
			fileType = "File"
			color = green
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
	maxName := 4  // "NAME"
	maxSize := 4  // "SIZE"
	maxType := 4  // "TYPE"
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
	fmt.Printf("%s╭%s┬%s┬%s┬%s╮%s\n",
		gray,
		strings.Repeat("─", maxName+4),
		strings.Repeat("─", maxSize+2),
		strings.Repeat("─", maxType+2),
		strings.Repeat("─", maxPerm+2),
		reset)

	fmt.Printf("%s│%s %s%-*s%s %s│%s %s%-*s%s %s│%s %s%-*s%s %s│%s %s%-*s%s %s│%s\n",
		gray, reset,
		yellow, maxName+2, "NAME", reset,
		gray, reset,
		yellow, maxSize, "SIZE", reset,
		gray, reset,
		yellow, maxType, "TYPE", reset,
		gray, reset,
		yellow, maxPerm, "PERMISSIONS", reset,
		gray, reset)

	fmt.Printf("%s├%s┼%s┼%s┼%s┤%s\n",
		gray,
		strings.Repeat("─", maxName+4),
		strings.Repeat("─", maxSize+2),
		strings.Repeat("─", maxType+2),
		strings.Repeat("─", maxPerm+2),
		reset)

	// Print files
	for _, f := range files {
		sizeStr := formatSize(f.size)
		fmt.Printf("%s│%s %s%s %s%-*s%s %s│%s %-*s %s│%s %-*s %s│%s %-*s %s│%s\n",
			gray, reset,
			f.color, f.icon, reset,
			maxName, f.name,
			reset,
			gray, reset,
			maxSize, sizeStr,
			gray, reset,
			maxType, f.fileType,
			gray, reset,
			maxPerm, f.permissions,
			gray, reset)
	}

	// Print footer
	fmt.Printf("%s╰%s┴%s┴%s┴%s╯%s\n",
		gray,
		strings.Repeat("─", maxName+4),
		strings.Repeat("─", maxSize+2),
		strings.Repeat("─", maxType+2),
		strings.Repeat("─", maxPerm+2),
		reset)
}
