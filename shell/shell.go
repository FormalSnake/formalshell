package shell

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
)

// ExecuteCommand runs an external command.
func ExecuteCommand(command string, args []string, customPath string) {
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

// HandlePipes splits a command by pipes (`|`) and sets up a pipeline.
func HandlePipes(input string) {
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

// LoadConfig loads shell configuration from various sources
func LoadConfig() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	customPath := os.Getenv("PATH")

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

	return customPath, nil
}
