package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestSessionStorage(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config", "session.json")
	if _, err := LoadSession(path); err == nil {
		t.Fatal("missing session must require login")
	}
	if err := SaveSession(path, "first-session"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { DeleteSession(path) })
	if useKeychain {
		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Fatal("Keychain credential written to disk")
		}
		if err := SaveSession(path, "replacement-session"); err != nil {
			t.Fatal(err)
		}
		value, err := LoadSession(path)
		if err != nil || value != "replacement-session" {
			t.Fatal("Keychain round trip failed", err)
		}
		if err := DeleteSession(path); err != nil {
			t.Fatal(err)
		}
		if _, err := LoadSession(path); err == nil {
			t.Fatal("Keychain credential survived logout")
		}
		return
	}
	if err := os.Chmod(path, 0644); err != nil {
		t.Fatal(err)
	}
	if err := SaveSession(path, "replacement-session"); err != nil {
		t.Fatal(err)
	}
	session, err := LoadSession(path)
	if err != nil || session != "replacement-session" {
		t.Fatalf("session round trip failed: %v", err)
	}
	info, err := os.Stat(path)
	if err != nil || info.Mode().Perm() != 0600 {
		t.Fatalf("session permissions: %v, %v", info, err)
	}
	if err := SaveSession(path, ""); err == nil {
		t.Fatal("empty session accepted")
	}
	if err := os.WriteFile(path, []byte(`{"session":""}`), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := LoadSession(path); err == nil {
		t.Fatal("empty stored session accepted")
	}
}
