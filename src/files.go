package main

import (
	"fmt"
	"os"
)

// readFile reads the content of a file and returns it as a string
// Returns empty string if file doesn't exist or can't be read
func readFile(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return string(data)
}

// writeFile writes content to a file with standard permissions
func writeFile(path, content string) error {
	return os.WriteFile(path, []byte(content), FilePermissions)
}

// readFileWithLimit reads a file with size validation
func readFileWithLimit(path string, maxSize int64) (string, error) {
	info, err := os.Stat(path)
	if err != nil {
		return "", err
	}
	if info.Size() > maxSize {
		return "", fmt.Errorf("file too large: %d bytes", info.Size())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}
