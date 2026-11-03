//go:build !darwin

package auth

const useKeychain = false

func keychainLoad(string) (string, error) { panic("Keychain requires macOS") }
func keychainSave(string, string) error   { panic("Keychain requires macOS") }
func keychainDelete(string) error         { panic("Keychain requires macOS") }
