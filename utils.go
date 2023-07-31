package main

import "strings"

func isHiddenDir(name string) bool {
	// make sure it's not the current directory first
	return name != "." && strings.Index(name, ".") == 0
}
