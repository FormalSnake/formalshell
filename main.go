package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"formalshell/cmds"
	"github.com/chzyer/readline"
)

// commandHistory stores unique commands that have been executed
var commandHistory = make(map[string]bool)

// displayPrompt generates the shell prompt, showing only the current folder name.
func displayPrompt() string {
	wd, err := os.Getwd()
	if err != nil {
		wd = "unknown"
	}

	blue := "\033[34m"
	reset := "\033[0m"
	return fmt.Sprintf("󰅟  %s%s%s > ", blue, filepath.Base(wd), reset)
}

// handleInput processes user input, including pipes and command chaining.
func handleInput(input string) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Add command to history
	parts := strings.Fields(input)
	if len(parts) > 0 {
		commandHistory[parts[0]] = true
	}

	// Split by `&&` for command chaining
	commands := strings.Split(input, "&&")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)

		// Handle pipes (`|`)
		if strings.Contains(cmd, "|") {
			handlePipes(cmd)
		} else {
			handleCommand(cmd)
		}
	}
}

// handleCommand processes a single command without pipes.
func handleCommand(input string) {
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	// Handle built-in commands
	switch command {
	case "exit":
		fmt.Println("Goodbye!")
		os.Exit(0)
	case "cd":
		cmds.HandleCD(args)
		return
	case "ls":
		cmds.CustomLS()
		return
	}

	// Execute external commands
	executeCommand(command, args)
}

// handlePipes splits a command by pipes (`|`) and sets up a pipeline.
func handlePipes(input string) {
	commands := strings.Split(input, "|")
	var prevCmd *exec.Cmd

	for i, cmdStr := range commands {
		cmdStr = strings.TrimSpace(cmdStr)
		parts := strings.Fields(cmdStr)

		if len(parts) == 0 {
			continue
		}

		cmd := exec.Command(parts[0], parts[1:]...)

		// Connect stdout of the previous command to stdin of the current one
		if prevCmd != nil {
			pipe, _ := prevCmd.StdoutPipe()
			cmd.Stdin = pipe
		} else {
			cmd.Stdin = os.Stdin
		}

		// Set stdout to os.Stdout for the last command
		if i == len(commands)-1 {
			cmd.Stdout = os.Stdout
		}

		cmd.Stderr = os.Stderr

		// Start the previous command before chaining
		if prevCmd != nil {
			_ = prevCmd.Start()
		}

		prevCmd = cmd
	}

	// Run the final command in the chain
	if prevCmd != nil {
		_ = prevCmd.Run()
	}
}

// executeCommand runs an external command.
func executeCommand(command string, args []string) {
	cmd := exec.Command(command, args...)
	cmd.Stdout = os.Stdout // Output to the shell's stdout
	cmd.Stderr = os.Stderr // Output errors to the shell's stderr
	cmd.Stdin = os.Stdin   // Allow input for commands like `cat`

	if err := cmd.Run(); err != nil {
		fmt.Printf("%s: command not found\n", command)
	}
}

// getDirCompletions returns a list of subdirectories for the given path
func getDirCompletions(path string) []string {
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
		searchDir = "."
		prefix = path
	}
	
	// List all directories
	if entries, err := os.ReadDir(searchDir); err == nil {
		for _, entry := range entries {
			name := entry.Name()
			if entry.IsDir() && strings.HasPrefix(name, prefix) {
				// Return just the completion part, not the full name
				completion := name[len(prefix):]
				if completion != "" {
					completions = append(completions, completion)
				}
			}
		}
	}
	
	return completions
}

// createCompleter returns a readline.PrefixCompleter based on command history and directory completion
func createCompleter() *readline.PrefixCompleter {
	var completions []readline.PrefixCompleterInterface
	
	// Add command history completions
	for cmd := range commandHistory {
		completions = append(completions, readline.PcItem(cmd))
	}
	
	// Add cd command with directory completion
	cdCompleter := readline.PcItem("cd",
		readline.PcItemDynamic(func(path string) []string {
			if path == "" {
				path = "."
			}
			return getDirCompletions(path)
		}))
	
	completions = append(completions, cdCompleter)
	
	return readline.NewPrefixCompleter(completions...)
}

func main() {
	// Configure readline
	config := &readline.Config{
		AutoComplete:          createCompleter(),
		InterruptPrompt:       "^C",
		EOFPrompt:            "exit",
		DisableAutoSaveHistory: false,  // Enable auto-save history
		HistorySearchFold:      true,
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	for {
		rl.SetPrompt(displayPrompt())
		line, err := rl.Readline()
		if err != nil { // io.EOF, readline.ErrInterrupt
			break
		}
		handleInput(line)
		
		// Update completer with new history and save to history
		rl.Config.AutoComplete = createCompleter()
		rl.SaveHistory(line)
	}

	fmt.Println("Shell exited.")
}
