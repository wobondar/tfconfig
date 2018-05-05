package main

import (
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var pwd, _ = os.Getwd()

type App struct {
	cli         *kingpin.Application
	args        []string
	pwd         string
	log         *Log
	isCi        bool
	projectPath string
}

func Init() (a *App) {
	a = &App{
		cli:  kingpin.New("tfconfig", "Terraform configuration manager"),
		args: os.Args[1:],
		pwd:  pwd,
	}

	a.log = a.Logger()

	a.cli.UsageWriter(a.log.ioWriter).ErrorWriter(a.log.ioWriter)

	a.cli.Version(Version)
	a.cli.HelpFlag.Short('h')
	a.cli.VersionFlag.Short('v')

	a.cli.Flag("ci", "CI flag, default 'false', if 'true' that you will not be asked before changes").
		Default("false").
		Short('c').
		Envar(CiEnvVar).
		BoolVar(&a.isCi)

	a.cli.Flag("path", "Terraform project path").
		Default(a.pwd).
		Short('p').
		PlaceHolder("PATH").
		StringVar(&a.projectPath)

	a.cli.Flag("verbose", "Verbose mode, default 'false'").
		Default("false").
		Short('V').
		BoolVar(&a.log.verbose)

	ConfigureEnvCommand(a)
	ConfigureDotEnvCommand(a)
	ConfigureBackendCommand(a)

	kingpin.MustParse(a.cli.Parse(a.args))

	return a
}
