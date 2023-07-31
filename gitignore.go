package main

import (
	"os"
	"strings"
)

var gitIgnoreFiles []string = nil

func getGitIgnoreFiles() []string {
	if gitIgnoreFiles == nil {
		b, err := os.ReadFile(".gitignore")
		if err == nil {
			gitIgnoreFiles = strings.Split(string(b), "\n")
		}
	}
	return gitIgnoreFiles
}

func isInGitIgnore(filename string) bool {
	for _, gitIgnoreFile := range getGitIgnoreFiles() {
		if strings.Contains(gitIgnoreFile, filename) {
			return true
		}
	}
	return false
}
