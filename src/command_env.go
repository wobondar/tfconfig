package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

type EnvCommand struct {
	app           *App
	log           *Log
	environment   string
	modulesPath   string
	modulesSource string
}

func ConfigureEnvCommand(a *App) {
	c := &EnvCommand{
		app: a,
		log: a.log,
	}
	cmd := a.cli.Command("env", "Switch Terraform project environment").PreAction(c.validate).Action(c.run)
	cmd.Arg("environment", "Environment name").Required().Envar(TerraformEnvVar).StringVar(&c.environment)
}

func (c *EnvCommand) run(context *kingpin.ParseContext) error {
	c.modulesSource = GetFullPath(c.modulesPath, "environment", c.environment, ConfigModuleName)
	c.log.Info("Module source will be: '%s'", c.modulesSource)

	c.app.AskConfirmOrSkip(c.app.isCi)

	if isCreated := c.createOrPopulateEnvironment(c.app.projectPath); isCreated {
		c.log.Info("Environment successfully switched: %s", c.environment)
	} else {
		c.log.ErrorF("I don't really know what exactly should be happen to cause that error ¯\\_(ツ)_/¯ ")
	}

	return nil
}

func (c *EnvCommand) validate(context *kingpin.ParseContext) error {
	c.app.ValidatePath()

	configFilePath := GetFullPath(c.app.projectPath, ConfigFile)

	c.log.ShowOpts("Config", configFilePath)
	if isExists, _ := ValidateFile(configFilePath); !isExists {
		c.log.ErrorFWithUsage("Configuration file '%s' does'nt exists", ConfigFile)
	}

	environmentConfig := GetFullPath(c.app.projectPath, EnvironmentFile)
	c.log.ShowOpts("Environment", environmentConfig)

	if err, isValid := ValidateEnvironment(c.environment); !isValid {
		c.log.ErrorFWithUsage(err)
	}

	if isExists, isWritable := ValidateFile(environmentConfig); isExists && isWritable {
		c.log.Warning("Environment file '%s' exists and will be overridden", EnvironmentFile)
	} else if isExists && !isWritable {
		c.log.ErrorF("Environment file '%s' exists, but dont have write permissions", EnvironmentFile)
	} else {
		c.log.Info("Environment file '%s' does'nt exists and will be created", EnvironmentFile)
	}

	modulesPath, isFoundModules := c.findModules(c.app.projectPath)
	if !isFoundModules {
		c.log.ErrorF("Cant find '%s' dir", ModulesFolder)
	}

	c.modulesPath = modulesPath
	return nil
}

func (c *EnvCommand) findModules(path string) (modulesPath string, isFound bool) {
	for _, v := range listSearchPaths() {
		if c.app.FindFolder(GetFullPath(path, v), ModulesFolder) {
			return v + ModulesFolder, true
		}
	}
	return "", false
}

func (c *EnvCommand) createOrPopulateEnvironment(projectPath string) (isCreated bool) {
	filePath := GetFullPath(projectPath, EnvironmentFile)
	if isExists, _ := ValidateFile(filePath); isExists {
		// already exists, replace
		c.app.ReplaceFile(filePath, c.generateEnvironmentConfigSource())
		return true
	} else {
		// does'nt exits, create
		c.app.CreateFile(filePath, c.generateEnvironmentConfigSource())
		return true
	}
	return false
}

func (c *EnvCommand) generateEnvironmentConfigSource() string {
	return WarningHeader +
		fmt.Sprintf("module \"%s\" {\n", ConfigModuleName) +
		fmt.Sprintf("  source = \"%s\"\n}\n\n", c.modulesSource)
}

func listSearchPaths() (paths []string) {
	return []string{"./", "../", "../../", "../../../", "../../../../"}
}
