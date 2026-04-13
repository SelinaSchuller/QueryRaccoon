package storage

import (
	"QueryRaccoon/internal/drivers"
	"encoding/json"
	"os"
	"path/filepath"
)

type SavedConnection struct {
	ID     string                   `json:"id"`
	Name   string                   `json:"name"`
	Config drivers.ConnectionConfig `json:"config"`
}

func configPath() (string, error) {
	dir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "QueryRaccoon", "connections.json"), nil
}

func Load() ([]SavedConnection, error) {
	path, err := configPath()
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []SavedConnection{}, nil
	}
	if err != nil {
		return nil, err
	}
	var conns []SavedConnection
	if err := json.Unmarshal(data, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}

func Save(conns []SavedConnection) error {
	path, err := configPath()
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(conns, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadDevSeed reads .devconnections.json from the current working directory.
// If the file doesn't exist, returns nil with no error.
func LoadDevSeed() ([]SavedConnection, error) {
	data, err := os.ReadFile(".devconnections.json")
	if os.IsNotExist(err) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var conns []SavedConnection
	if err := json.Unmarshal(data, &conns); err != nil {
		return nil, err
	}
	return conns, nil
}
