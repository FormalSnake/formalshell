package cmds

import (
	"path/filepath"
	"strings"
)

// FileIcons maps file extensions and types to Nerd Font icons
var FileIcons = map[string]string{
	// Folders
	"folder":          "󰉋",  // Default folder
	"folder_open":     "󰝰",  // Open folder
	"folder_config":   "󱁿",  // Config folder
	"folder_git":      "󰊢",  // Git folder
	"folder_github":   "󰊤",  // GitHub folder
	"folder_home":     "󱂵",  // Home folder
	"folder_docs":     "󰈙",  // Documents folder
	"folder_images":   "󰉏",  // Images folder
	"folder_music":    "󱍙",  // Music folder
	"folder_videos":   "󰕧",  // Videos folder
	"folder_downloads": "󰇚", // Downloads folder

	// Development
	".go":           "󰟓",   // Go files
	".py":           "󰌠",   // Python files
	".js":           "󰌞",   // JavaScript files
	".ts":           "󰛦",   // TypeScript files
	".jsx":          "󰜈",   // React files
	".tsx":          "󰜈",   // React TypeScript files
	".vue":          "󰡄",   // Vue files
	".rs":           "󱘗",   // Rust files
	".cpp":          "󰙲",   // C++ files
	".c":            "󰙱",   // C files
	".h":            "󰙲",   // Header files
	".java":         "󰬷",   // Java files
	".kt":           "󱈙",   // Kotlin files
	".rb":           "󰴭",   // Ruby files
	".php":          "󰌟",   // PHP files
	".scala":        "󰘜",   // Scala files
	".swift":        "󰛥",   // Swift files

	// Web
	".html":         "󰌝",   // HTML files
	".css":          "󰌜",   // CSS files
	".scss":         "󰌜",   // SCSS files
	".sass":         "󰌜",   // Sass files
	".json":         "󰘦",   // JSON files
	".xml":          "󰗀",   // XML files
	".yaml":         "󰈙",   // YAML files
	".yml":          "󰈙",   // YML files
	".md":           "󰍔",   // Markdown files
	".txt":          "󰈙",   // Text files

	// Data & Databases
	".sql":          "󰆼",   // SQL files
	".db":           "󰆼",   // Database files
	".csv":          "󰈛",   // CSV files
	".xlsx":         "󰈛",   // Excel files
	".doc":          "󰈬",   // Word files
	".pdf":          "󰈦",   // PDF files

	// Media
	".mp3":          "󰎆",   // Audio files
	".wav":          "󰎆",   // Wave files
	".mp4":          "󰕧",   // Video files
	".mov":          "󰕧",   // Movie files
	".png":          "󰋩",   // PNG files
	".jpg":          "󰋩",   // JPG files
	".jpeg":         "󰋩",   // JPEG files
	".gif":          "󰋩",   // GIF files
	".svg":          "󰜡",   // SVG files
	".ico":          "󰀲",   // Icon files

	// System
	"executable":    "󰆍",   // Executable files
	"symlink":       "󰉒",   // Symbolic links
	"file":          "󰈙",   // Default file
	".sh":           "󰆍",   // Shell scripts
	".bash":         "󰆍",   // Bash scripts
	".zsh":          "󰆍",   // Zsh scripts
	".vim":          "󰕷",   // Vim files
	".nvim":         "󰕷",   // Neovim files
	".gitignore":    "󰈉",   // Git ignore files
	".dockerignore": "󰡨",   // Docker ignore files
	"dockerfile":    "󰡨",   // Dockerfiles
	".env":          "󰒓",   // Environment files
	".log":          "󰌱",   // Log files
	".lock":         "󰌾",   // Lock files
	".zip":          "󰗄",   // Zip files
	".tar":          "󰗄",   // Tar files
	".gz":           "󰗄",   // Gzip files
	".7z":           "󰗄",   // 7zip files
	".iso":          "󰗮",   // ISO files
}

// GetFileIcon returns the appropriate icon for a file based on its name and type
func GetFileIcon(name string, isDir bool, isExecutable bool, isSymlink bool) string {
	// Check for special cases first
	if isSymlink {
		return FileIcons["symlink"]
	}
	if isDir {
		// Special folder cases
		switch name {
		case ".git":
			return "󰊢"  // Git folder icon
		case ".github":
			return FileIcons["folder_github"]
		case "config", ".config":
			return FileIcons["folder_config"]
		case "home":
			return FileIcons["folder_home"]
		case "Documents", "docs":
			return FileIcons["folder_docs"]
		case "Pictures", "images":
			return FileIcons["folder_images"]
		case "Music":
			return FileIcons["folder_music"]
		case "Videos":
			return FileIcons["folder_videos"]
		case "Downloads":
			return FileIcons["folder_downloads"]
		default:
			return FileIcons["folder"]
		}
	}
	if isExecutable {
		return FileIcons["executable"]
	}

	// Check for exact filename matches
	lowerName := strings.ToLower(name)
	switch lowerName {
	case "dockerfile":
		return FileIcons["dockerfile"]
	case ".gitignore":
		return FileIcons["gitignore"]
	case ".dockerignore":
		return FileIcons["dockerignore"]
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(name))
	if icon, ok := FileIcons[ext]; ok {
		return icon
	}

	// Default file icon
	return FileIcons["file"]
}
