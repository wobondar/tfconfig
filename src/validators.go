package main

import (
	"golang.org/x/sys/unix"
	"os"
	"strings"
	"fmt"
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
	restrictedCharacters := " -_\\/~!@#$^&*)(=+`][;:'\"?><}{,."
	if strings.ContainsAny(environment, restrictedCharacters) {
		return fmt.Sprintf("Environment name cant contain characters:\t %s", restrictedCharacters), false
	}

	return "", true
}

func (a *App) ValidatePath() error {
	a.log.ShowOpts("Path", a.projectPath)
	if err, isValid := ValidatePath(a.projectPath); !isValid {
		a.log.ErrorFWithUsage(err)
	}

	return nil
}
