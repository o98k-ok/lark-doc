package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"

	"github.com/o98k-ok/aggregation/auth"
	"github.com/o98k-ok/aggregation/doc"
)

func cachePath() (string, error) {
	directory, err := os.UserCacheDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(directory, "lark-doc", "documents.json"), nil
}

func saveCache(path string, entities []doc.Entity) error {
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".documents-*")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	err = json.NewEncoder(file).Encode(entities)
	closeErr := file.Close()
	if err != nil {
		return err
	}
	if closeErr != nil {
		return closeErr
	}
	return os.Rename(file.Name(), path)
}

func logout() error {
	path, err := auth.SessionPath()
	if err != nil {
		return err
	}
	sessionErr := auth.DeleteSession(path)
	cache, err := cachePath()
	if err != nil {
		return errors.Join(sessionErr, err)
	}
	cacheErr := os.Remove(cache)
	if os.IsNotExist(cacheErr) {
		cacheErr = nil
	}
	return errors.Join(sessionErr, cacheErr)
}

func loginItem() {
	// Actions are separate from document URLs; the workflow dispatcher handles them.
	json.NewEncoder(os.Stdout).Encode(map[string]interface{}{"items": []interface{}{
		map[string]interface{}{"title": "登录飞书", "subtitle": "回车打开本机终端，使用飞书扫码；凭证保存在 Keychain", "arg": "lark-doc:login", "valid": true},
	}})
}
