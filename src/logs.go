package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"io"
	"os"
	"strings"
)

type Log struct {
	cli      *kingpin.Application
	args     []string
	ioWriter io.Writer
	verbose  bool
	silent   bool
	isQuite  bool
	awsDebug bool
}

func (a *App) Logger() *Log {
	return &Log{
		cli:      a.cli,
		args:     a.args,
		verbose:  false,
		silent:   false,
		isQuite:  false,
		awsDebug: false,
		ioWriter: os.Stdout,
	}
}

func (l *Log) HandleSilent() {
	if l.silent {
		l.Quite()
	}
}

func (l *Log) EnableAwsDebug() {
	l.awsDebug = true
}

func (l *Log) Quite() {
	if !l.verbose {
		l.isQuite = true
	}
	l.ioWriter = os.Stderr
	l.cli.UsageWriter(l.ioWriter).ErrorWriter(l.ioWriter)
}

func (l *Log) ErrorF(format string, s ...interface{}) {
	showLog("ERROR", fmt.Sprintf(format, s...), true, false, l.ioWriter)
}

func (l *Log) Info(format string, s ...interface{}) {
	showLog("INFO", fmt.Sprintf(format, s...), false, l.isQuite, l.ioWriter)
}

func (l *Log) Debug(format string, s ...interface{}) {
	if l.verbose {
		showLog("DEBUG", fmt.Sprintf(format, s...), false, l.isQuite, l.ioWriter)
	}
}

func (l *Log) Warning(format string, s ...interface{}) {
	showLog("WARNING", fmt.Sprintf(format, s...), false, l.isQuite, l.ioWriter)
}

func (l *Log) ShowOpts(name string, value string) {
	showLog("INFO", fmt.Sprintf("%s:\t%s", name, value), false, l.isQuite, l.ioWriter)
}

func (l *Log) Printf(format string, s ...interface{}) {
	fmt.Fprintf(os.Stdout, format, s...)
}

func (l *Log) must(err error) {
	if err != nil {
		l.ErrorF("Message: %v", err)
	}
}

func showLog(level string, message string, exit bool, quite bool, wr io.Writer) error {
	if !quite {
		fmt.Fprintf(wr, "[%s]  %s\n", strings.ToUpper(level), message)
	}
	if exit {
		os.Exit(1)
	}
	return nil
}

func (l *Log) ErrorFWithUsage(format string, s ...interface{}) error {
	l.Usage()
	l.ErrorF(format, s...)
	return nil
}

func (l *Log) Usage() {
	l.cli.Usage(l.args)
}
