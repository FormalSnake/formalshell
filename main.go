package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"formalshell/cmds"
	"formalshell/completions"
	"formalshell/history"
	"formalshell/shell"
	"github.com/chzyer/readline"
)

// Global state
var (
	aliases    = make(map[string]string)
	customPath string
)


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
func handleInput(input string, rl *readline.Instance, hist *history.History) {
	input = strings.TrimSpace(input)
	if input == "" {
		return
	}

	// Save the raw input to history immediately
	if input = strings.TrimSpace(input); input != "" {
		hist.CommandHistory[input] = true
		if err := hist.Save(); err != nil {
			fmt.Printf("Error saving history: %v\n", err)
		}
	}

	// Process the commands
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
	
	// Check aliases first
	if alias, exists := aliases[command]; exists {
		// Replace the command with its alias
		aliasParts := strings.Fields(alias)
		command = aliasParts[0]
		args := append(aliasParts[1:], parts[1:]...)
		parts = append([]string{command}, args...)
	}
	
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
		cmds.CustomLS(args...)
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
	// Set up signal handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT)
	defer signal.Stop(sigChan)

	// Create a shell script that will execute the command
	script := fmt.Sprintf(". ~/.config/formalshell/config\n%s %s", command, strings.Join(args, " "))
	
	tmpFile, err := os.CreateTemp("", "formalsh_cmd_*.sh")
	if err != nil {
		fmt.Printf("Error creating temp file: %v\n", err)
		return
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.WriteString(script); err != nil {
		fmt.Printf("Error writing temp file: %v\n", err)
		return
	}
	tmpFile.Close()

	// Execute the command in a shell context
	cmd := exec.Command("/bin/sh", tmpFile.Name())
	cmd.Env = append(os.Environ(), "PATH="+customPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	// Start command in background
	if err := cmd.Start(); err != nil {
		fmt.Printf("%s: command not found\n", command)
		return
	}

	// Handle command interruption
	go func() {
		select {
		case <-sigChan:
			if cmd.Process != nil {
				cmd.Process.Signal(syscall.SIGINT)
			}
		}
	}()

	// Wait for command completion
	err = cmd.Wait()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Don't print status for interrupted commands
			if exitErr.ExitCode() != int(syscall.SIGINT) {
				fmt.Printf("Command exited with status %d\n", exitErr.ExitCode())
			}
		}
		return
	}
}

func loadConfig() {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return
	}

	// First source system profile and zshenv to get basic PATH and env vars
	script := `
		. /etc/profile
		[ -f ~/.zshenv ] && . ~/.zshenv
		env > "$TMPDIR/formalsh_env"
	`
	tmpFile, err := os.CreateTemp("", "formalsh_*.sh")
	if err == nil {
		defer os.Remove(tmpFile.Name())
		if _, err := tmpFile.WriteString(script); err == nil {
			tmpFile.Close()
			cmd := exec.Command("/bin/sh", tmpFile.Name())
			cmd.Run()

			// Read and apply environment
			if envData, err := os.ReadFile(os.Getenv("TMPDIR") + "/formalsh_env"); err == nil {
				envVars := strings.Split(string(envData), "\n")
				for _, line := range envVars {
					if parts := strings.SplitN(line, "=", 2); len(parts) == 2 {
						os.Setenv(parts[0], parts[1])
						if parts[0] == "PATH" {
							customPath = parts[1]
						}
					}
				}
				os.Remove(os.Getenv("TMPDIR") + "/formalsh_env")
			}
		}
	}

	// Then load our specific config
	configPath := filepath.Join(homeDir, ".config", "formalshell", "config")
	if _, err := os.Stat(configPath); err == nil {
		cmd := exec.Command("/bin/sh", configPath)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = os.Environ()
		if err := cmd.Run(); err != nil {
			fmt.Printf("Error loading formalsh config: %v\n", err)
		}
	}
}

func main() {
	// Load config file before starting shell
	var err error
	customPath, err = shell.LoadConfig()
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		os.Exit(1)
	}

	// Initialize history
	hist, err := history.New()
	if err != nil {
		fmt.Printf("Error initializing history: %v\n", err)
		os.Exit(1)
	}

	// Configure readline
	config := &readline.Config{
		AutoComplete:           completions.CreateCompleter(hist.CommandHistory),
		InterruptPrompt:        "^C",
		EOFPrompt:             "exit",
		DisableAutoSaveHistory: false,
		HistorySearchFold:      true,
	}

	instance, err := readline.NewEx(config)
	if err != nil {
		panic(err)
	}
	defer instance.Close()

	// Load command history
	if err := hist.Load(instance); err != nil {
		fmt.Printf("Error loading history: %v\n", err)
	}
	defer hist.Save()

	for {
		instance.SetPrompt(displayPrompt())
		line, err := instance.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				fmt.Println()
				continue
			} else if err == io.EOF {
				break
			}
			fmt.Printf("Error: %v\n", err)
			continue
		}
		handleInput(line, instance, hist)

		// Save history but keep the existing completer
		instance.SaveHistory(line)
	}

	fmt.Println("Shell exited.")
}
