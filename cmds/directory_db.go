package cmds

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

type DirectoryEntry struct {
	Path      string    `json:"path"`
	Score     float64   `json:"score"`
	LastVisit time.Time `json:"last_visit"`
}

type DirectoryDB struct {
	Entries []DirectoryEntry `json:"entries"`
	dbPath  string
}

func NewDirectoryDB() *DirectoryDB {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return &DirectoryDB{}
	}

	dbDir := filepath.Join(homeDir, ".config", "formalshell")
	if err := os.MkdirAll(dbDir, 0755); err != nil {
		return &DirectoryDB{}
	}

	dbPath := filepath.Join(dbDir, "directory.json")
	db := &DirectoryDB{dbPath: dbPath}
	db.load()
	return db
}

func (db *DirectoryDB) load() {
	data, err := os.ReadFile(db.dbPath)
	if err != nil {
		return
	}
	json.Unmarshal(data, &db)
}

func (db *DirectoryDB) save() {
	data, err := json.MarshalIndent(db, "", "  ")
	if err != nil {
		return
	}
	os.WriteFile(db.dbPath, data, 0644)
}

func (db *DirectoryDB) AddVisit(path string) {
	path = filepath.Clean(path)
	
	// Update existing entry or add new one
	for i := range db.Entries {
		if db.Entries[i].Path == path {
			db.Entries[i].Score += 1
			db.Entries[i].LastVisit = time.Now()
			db.save()
			return
		}
	}

	// Add new entry
	db.Entries = append(db.Entries, DirectoryEntry{
		Path:      path,
		Score:     1,
		LastVisit: time.Now(),
	})
	db.save()
}

func (db *DirectoryDB) FindMatch(partial string) string {
	if len(db.Entries) == 0 {
		return partial
	}

	// Sort entries by score
	sort.Slice(db.Entries, func(i, j int) bool {
		return db.Entries[i].Score > db.Entries[j].Score
	})

	// Try exact match first
	for _, entry := range db.Entries {
		if strings.Contains(filepath.Base(entry.Path), partial) {
			return entry.Path
		}
	}

	// Try fuzzy match
	for _, entry := range db.Entries {
		if strings.Contains(entry.Path, partial) {
			return entry.Path
		}
	}

	return partial
}
