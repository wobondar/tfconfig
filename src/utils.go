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
func (a *App) AskConfirmOrSkip(trigger bool) {
	if !trigger {
		var approved string
		fmt.Println("\nAfter this operation configuration will be changed")
		fmt.Printf("Do you want to continue? [Y/n] ")
		fmt.Scanln(&approved)
		if !strings.EqualFold(strings.ToLower(approved), "y") {
			a.log.ErrorF("Abort.")
		}
	} else {
		a.log.Warning("Confirmation has been skipped via running environment configuration")
	}
}

func (a *App) createOrPopulateFile(filePath string, content string) (isCreated bool) {
	if isExists, _ := ValidateFile(filePath); isExists {
		// already exists, replace
		a.ReplaceFile(filePath, content)
		return true
	} else {
		// does'nt exits, create
		a.CreateFile(filePath, content)
		return true
	}
	return false
}

func (a *App) CreateFile(filePath string, content string) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if a.isError(err) {
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if a.isError(err) {
		return
	}

	err = file.Sync()
	if a.isError(err) {
		return
	}
}

func (a *App) ReplaceFile(filePath string, content string) {
	file, err := os.OpenFile(filePath, os.O_RDWR, 0644)
	if a.isError(err) {
		return
	}
	defer file.Close()

	_, err = file.WriteString(content)
	if a.isError(err) {
		return
	}

	err = file.Sync()
	if a.isError(err) {
		return
	}
}

func (a *App) FindFolder(path string, dir string) (isFound bool) {
	file, err := os.Open(path)
	if a.isError(err) {
		return
	}
	defer file.Close()

	a.log.Info("Looking in '%s'", path)

	fileList, _ := file.Readdir(0)
	for _, files := range fileList {
		// entity must be equal to our modules folder name and must be exactly the directory or can be as SymLink
		if strings.EqualFold(files.Name(), dir) && (files.IsDir() || files.Mode()&os.ModeSymlink != 0) {
			a.log.Info("Found '%s' in '%s'", dir, path)
			return true
		}
	}
	return false
}

func (a *App) isError(err error) bool {
	if err != nil {
		a.log.ErrorF("IO error: %s", err.Error())
	}

	return err != nil
}
