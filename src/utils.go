package main

import (
	"fmt"
	"os"
	"strings"
)

const pathSeparator = "/"

func GetFullPath(parts ...string) string {
	var fullPath string
	first := true
	for _, v := range parts {
		if first {
			first = false
			fullPath = v
		} else {
			fullPath = fullPath + pathSeparator + v
		}
	}
	return fullPath
}

// trigger: true, confirmation will be skipped
// trigger: false, user will be asked to confirm changes
func AskConfirmOrSkip(trigger bool) {
	if !trigger {
		var approved string
		fmt.Println("\nAfter this operation configuration will be changed")
		fmt.Printf("Do you want to continue? [Y/n] ")
		fmt.Scanln(&approved)
		if !strings.EqualFold(strings.ToLower(approved), "y") {
			ShowError("Abort.")
		}
	} else {
		ShowWarning("Confirmation has been skipped via running environment configuration")
	}
}

func CreateFile(filePath string, content string) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if isError(err) {
		return
	}

	err = file.Sync()
	if isError(err) {
		return
	}
}

func ReplaceFile(filePath string, content string) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if isError(err) {
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if isError(err) {
		return
	}

	err = file.Sync()
	if isError(err) {
		return
	}
}

func FindFolder(path string, dir string) (isFound bool) {
	file, err := os.Open(path)
	if isError(err) {
		return
	}
	defer file.Close()

	ShowInfo("Looking in '%s'", path)

	fileList, _ := file.Readdir(0)
	for _, files := range fileList {
		// entity must be equal to our modules folder name and must be exactly the directory or can be as SymLink
		if strings.EqualFold(files.Name(), dir) && (files.IsDir() || files.Mode()&os.ModeSymlink != 0) {
			ShowInfo("Found '%s' in '%s'", dir, path)
			return true
		}
	}
	return false
}

func isError(err error) bool {
	if err != nil {
		ShowError("IO error: %s", err.Error())
	}

	return err != nil
}
