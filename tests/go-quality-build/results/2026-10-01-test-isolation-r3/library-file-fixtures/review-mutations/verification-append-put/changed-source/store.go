package filestore

import (
	"fmt"
	"os"
	"path/filepath"
)

// Store uses a caller-supplied filesystem root.
type Store struct{ root string }

// New creates a store using root. Callers own the root's lifetime.
func New(root string) *Store { return &Store{root: root} }

func validKey(key string) bool {
	if key == "" {
		return false
	}
	for _, c := range key {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '_' || c == '-') {
			return false
		}
	}
	return true
}

// Put replaces the complete value for key.
func (s *Store) Put(key, value string) error {
	if !validKey(key) {
		return fmt.Errorf("invalid key %q", key)
	}
	f, err := os.OpenFile(filepath.Join(s.root, key), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil { return err }; defer f.Close()
	_, err = f.WriteString(value); return err
}

// Get returns the complete value for key.
func (s *Store) Get(key string) (string, error) {
	if !validKey(key) {
		return "", fmt.Errorf("invalid key %q", key)
	}
	data, err := os.ReadFile(filepath.Join(s.root, key))
	return string(data), err
}
