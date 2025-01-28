package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"formalshell/cmds"
	"formalshell/completions"
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


func main() {
	// Configure readline
	config := &readline.Config{
		AutoComplete:          completions.CreateCompleter(commandHistory),
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
		rl.Config.AutoComplete = completions.CreateCompleter(commandHistory)
		rl.SaveHistory(line)
	}

	fmt.Println("Shell exited.")
}
