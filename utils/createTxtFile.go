package utils

import (
	"os"
	"strings"
)

// Buat fungsi reusable
func WriteToTXT(filename, content string) error {
	dir := getDirFromPath(filename)
	err := os.MkdirAll(dir, os.ModePerm) 
	if err != nil {
		return err
	}

	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// Ambil direktori dari path
func getDirFromPath(path string) string {
	lastSlash := strings.LastIndex(path, "/")
	if lastSlash == -1 {
		return "." 
	}
	return path[:lastSlash]
}
