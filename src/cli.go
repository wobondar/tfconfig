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
	isCi        bool
	projectPath string
}

func Init() (a *App) {
	a = &App{
		cli:  kingpin.New("tfconfig", "Terraform configuration manager"),
		args: os.Args[1:],
		pwd:  pwd,
	}

	a.cli.Version(Version)
	a.cli.HelpFlag.Short('h')
	a.cli.VersionFlag.Short('v')

	a.cli.Flag("ci", "CI flag, default 'false', if 'true' that you will not be asked before changes").Default("false").Short('c').Bool()
	a.cli.GetFlag("ci").Envar(CiEnvVar).BoolVar(&a.isCi)

	a.cli.Flag("path", "Terraform project path").PlaceHolder("PATH").Short('p').Default(a.pwd).String()
	a.cli.GetFlag("path").StringVar(&a.projectPath)

	a.cli.PreAction(a.validate)

	ConfigureEnvCommand(a)

	kingpin.MustParse(a.cli.Parse(a.args))

	return a
}

func (a *App) validate(context *kingpin.ParseContext) error {
	configFilePath := GetFullPath(a.projectPath, ConfigFile)

	ShowOpts("Path", a.projectPath)
	ShowOpts("Config", configFilePath)

	if err, isValid := ValidatePath(a.projectPath); !isValid {
		a.ShowErrorWithUsage(err)
	}

	if isExists, _ := ValidateFile(configFilePath); !isExists {
		a.ShowErrorWithUsage("Configuration file '%s' does'nt exists", ConfigFile)
	}

	return nil
}
