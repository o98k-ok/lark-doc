package auth

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

func SessionPath() (string, error) {
	directory, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "lark-doc", "session.json"), nil
}

func LoadSession(path string) (string, error) {
	if useKeychain {
		return keychainLoad(path)
	}
	contents, err := os.ReadFile(path)
	if err != nil {
		return "", errors.New("session unavailable; run lark-doc login")
	}
	var stored struct {
		Session string `json:"session"`
	}
	if json.Unmarshal(contents, &stored) != nil || stored.Session == "" {
		return "", errors.New("invalid session file; run lark-doc login")
	}
	return stored.Session, nil
}

func SaveSession(path, session string) error {
	if session == "" {
		return errors.New("cannot save an empty session")
	}
	if useKeychain {
		if err := keychainSave(path, session); err != nil {
			return err
		}
		return removeSessionFile(path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".session-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(struct {
		Session string `json:"session"`
	}{session})
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}

func DeleteSession(path string) error {
	if useKeychain {
		if err := keychainDelete(path); err != nil {
			return err
		}
	}
	return removeSessionFile(path)
}

func removeSessionFile(path string) error {
	err := os.Remove(path)
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
