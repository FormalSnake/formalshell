package history

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/chzyer/readline"
)

type History struct {
	CommandHistory map[string]bool
	HistoryFile   string
}

func New() (*History, error) {
	h := &History{
		CommandHistory: make(map[string]bool),
	}

	homeDir, err := os.UserHomeDir()
	if err == nil {
		historyDir := filepath.Join(homeDir, ".config", "formalshell")
		if err := os.MkdirAll(historyDir, 0755); err == nil {
			h.HistoryFile = filepath.Join(historyDir, "history")
		}
	}

	return h, nil
}

func (h *History) Load(rl *readline.Instance) error {
	if h.HistoryFile == "" {
		return nil
	}

	data, err := os.ReadFile(h.HistoryFile)
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
			h.CommandHistory[line] = true
		}
	}
	return nil
}

func (h *History) Save() error {
	if h.HistoryFile == "" {
		return nil
	}

	var lines []string
	for cmd := range h.CommandHistory {
		if cmd = strings.TrimSpace(cmd); cmd != "" {
			lines = append(lines, cmd)
		}
	}

	return os.WriteFile(h.HistoryFile, []byte(strings.Join(lines, "\n")+"\n"), 0666)
}
