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

// Global state
var (
	commandHistory = make(map[string]bool)
	aliases       = make(map[string]string)
	customPath    = os.Getenv("PATH")
	historyFile   string
)

func init() {
	homeDir, err := os.UserHomeDir()
	if err == nil {
		historyDir := filepath.Join(homeDir, ".config", "formalshell")
		if err := os.MkdirAll(historyDir, 0755); err == nil {
			historyFile = filepath.Join(historyDir, "history")
		}
	}
}

func loadHistory(rl *readline.Instance) error {
	if historyFile == "" {
		return nil
	}

	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line = strings.TrimSpace(line); line != "" {
			rl.SaveHistory(line)
		}
	}
	return nil
}

func saveHistory(rl *readline.Instance) error {
	if historyFile == "" {
		return nil
	}

	history := rl.GetHistory()
	var lines []string
	for _, item := range history.Deque {
		if str, ok := item.(string); ok {
			lines = append(lines, str)
		}
	}

	return os.WriteFile(historyFile, []byte(strings.Join(lines, "\n")+"\n"), 0666)
}

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
	// Create a shell script that will execute the command
	script := fmt.Sprintf(". ~/.config/.formalsh\n%s %s", command, strings.Join(args, " "))
	
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

	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			// Just print the error but don't exit the shell
			fmt.Printf("Command exited with status %d\n", exitErr.ExitCode())
		} else {
			fmt.Printf("%s: command not found\n", command)
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
	configPath := filepath.Join(homeDir, ".config", ".formalsh")
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
	loadConfig()

	// Configure readline
	config := &readline.Config{
		AutoComplete:           completions.CreateCompleter(commandHistory),
		InterruptPrompt:        "^C",
		EOFPrompt:              "exit",
		DisableAutoSaveHistory: false, // Enable auto-save history
		HistorySearchFold:      true,
	}

	rl, err := readline.NewEx(config)
	if err != nil {
		panic(err)
	}
	defer rl.Close()

	// Load command history
	if err := loadHistory(rl); err != nil {
		fmt.Printf("Error loading history: %v\n", err)
	}
	defer saveHistory(rl)

	for {
		rl.SetPrompt(displayPrompt())
		line, err := rl.Readline()
		if err != nil {
			if err == readline.ErrInterrupt {
				// For Ctrl+C, just continue the loop
				continue
			} else {
				// For other errors (like EOF from Ctrl+D), exit
				break
			}
		}
		handleInput(line)

		// Update completer with new history and save to history
		rl.Config.AutoComplete = completions.CreateCompleter(commandHistory)
		rl.SaveHistory(line)
	}

	fmt.Println("Shell exited.")
}
