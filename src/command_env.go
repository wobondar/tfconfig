package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
)

type EnvCommand struct {
	app           *App
	environment   string
	modulesPath   string
	modulesSource string
}

func ConfigureEnvCommand(a *App) {
	c := &EnvCommand{app: a}
	cmd := a.cli.Command("env", "Switch project environment").PreAction(c.validate).Action(c.run)
	cmd.Arg("environment", "Environment name").Required().Envar(TerraformEnvVar).StringVar(&c.environment)
}

func (c *EnvCommand) run(context *kingpin.ParseContext) error {
	c.modulesSource = GetFullPath(c.modulesPath, "environment", c.environment, ConfigModuleName)
	ShowInfo("Module source will be: '%s'", c.modulesSource)

	AskConfirmOrSkip(c.app.isCi)

	if isCreated := c.createOrPopulateEnvironment(c.app.projectPath); isCreated {
		ShowInfo("Environment successfully switched: %s", c.environment)
	} else {
		ShowError("I don't know what exactly should be happen to cause that error ¯\\_(ツ)_/¯ ")
	}

	return nil
}

func (c *EnvCommand) validate(context *kingpin.ParseContext) error {

	environmentConfig := GetFullPath(c.app.projectPath, EnvironmentFile)
	ShowOpts("Environment", environmentConfig)

	if err, isValid := ValidateEnvironment(c.environment); !isValid {
		c.app.ShowErrorWithUsage(err)
	}

	if isExists, isWritable := ValidateFile(environmentConfig); isExists && isWritable {
		ShowWarning("Environment file '%s' exists and will be overridden", EnvironmentFile)
	} else if isExists && !isWritable {
		ShowError("Environment file '%s' exists, but dont have write permissions", EnvironmentFile)
	} else {
		ShowInfo("Environment file '%s' does'nt exists and will be created", EnvironmentFile)
	}

	modulesPath, isFoundModules := findModules(c.app.projectPath)
	if !isFoundModules {
		ShowError("Cant find '%s' dir", ModulesFolder)
	}

	c.modulesPath = modulesPath
	return nil
}

func findModules(path string) (modulesPath string, isFound bool) {
	for _, v := range listSearchPaths() {
		if FindFolder(GetFullPath(path, v), ModulesFolder) {
			return v + ModulesFolder, true
		}
	}
	return "", false
}

func (c *EnvCommand) createOrPopulateEnvironment(projectPath string) (isCreated bool) {
	filePath := GetFullPath(projectPath, EnvironmentFile)
	if isExists, _ := ValidateFile(filePath); isExists {
		// already exists, replace
		ReplaceFile(filePath, c.generateEnvironmentConfigSource())
		return true
	} else {
		// does'nt exits, create
		CreateFile(filePath, c.generateEnvironmentConfigSource())
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
