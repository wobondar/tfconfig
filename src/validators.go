package main

import (
	"golang.org/x/sys/unix"
	"os"
	"strings"
)

func ValidateFile(file string) (isExists bool, isWritable bool) {
	if _, err := os.Stat(file); err == nil {
		return true, writable(file)
	}
	return false, false
}

func writable(path string) bool {
	return unix.Access(path, unix.W_OK) == nil
}

func ValidatePath(path string) (err string, isValid bool) {
	if path == "" {
		return "Project path cant be empty", false
	}
	if len(path) == 0 {
		return "Project path cant be empty", false
	}
	if !writable(path) {
		return "Project path must have write permissions", false
	}
	return "", true
}

func ValidateEnvironment(environment string) (err string, isValid bool) {
	if environment == "" {
		return "Environment name cant be empty", false
	}
	if len(environment) == 0 {
		return "Environment name cant be empty", false
	}
	if strings.ContainsAny(environment, " -_") {
		return "Environment name cant contain spaces, underscored and -", false
	}

	return "", true
}
