package main

import (
	"fmt"
	"os"
	"strings"
)

func ShowError(format string, a ...interface{}) {
	showLog("ERROR", fmt.Sprintf(format, a...), true)
}

func ShowInfo(format string, a ...interface{}) {
	showLog("INFO", fmt.Sprintf(format, a...), false)
}

func ShowOpts(name string, value string) {
	showLog("INFO", fmt.Sprintf("%s:\t%s", name, value), false)
}

func ShowWarning(format string, a ...interface{}) {
	showLog("WARNING", fmt.Sprintf(format, a...), false)
}

func showLog(level string, message string, exit bool) error {
	fmt.Printf("[%s]  %s\n", strings.ToUpper(level), message)
	if exit {
		os.Exit(1)
	}
	return nil
}

func (a *App) ShowErrorWithUsage(format string, s ...interface{}) error {
	a.ShowUsage()
	ShowError(format, s...)
	return nil
}

func (a *App) ShowUsage() {
	a.cli.Usage(a.args)
}
