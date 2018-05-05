package main

import (
	"bytes"
	"github.com/joho/godotenv"
	"gopkg.in/alecthomas/kingpin.v2"
	"path/filepath"
	"text/template"
)

const (
	// Template for generate backend config
	backendTemplate = `bucket = "{{.TerraformStateBucket}}"

key = "{{.Environment}}/{{.TerraformStateKey}}/terraform.tfstate"

region = "{{.Region}}"

dynamodb_table = "{{.TerraformLockTable}}"

kms_key_id = "{{.KmsKeyArn}}"
`

	// Default Terraform config file name which cant be overridden
	defaultTerraformBackendConfig = "terraform-backend.tfconf"

	// Global configuration that includes environment (dev, staging) configuration stored there
	defaultEnvironmentConfig = "environment.env"
	// Project specific configuration
	defaultProjectConfig = "terraform.env"
)

type BackendConfig struct {
	Environment          string
	Region               string
	TerraformStateBucket string
	TerraformStateKey    string
	TerraformLockTable   string
	KmsKeyArn            string
}

type dotEnv struct {
	environment map[string]string
	project     map[string]string
}

type BackendCommand struct {
	app                   *App
	log                   *Log
	environment           string
	modulesPath           string
	environmentConfigPath string
	projectConfigPath     string
	backendConfigPath     string
	backendConfig         *BackendConfig
	template              *template.Template
	dotEnvConfig          dotEnv
}

func ConfigureBackendCommand(a *App) {
	c := &BackendCommand{
		app: a,
		log: a.log,
	}
	cmd := a.cli.Command("backend", "Generate backend configuration").
		PreAction(c.validate).
		Action(c.run)

	cmd.Arg("environment", "Environment name").
		Required().
		StringVar(&c.environment)

	cmd.Flag("backend-config", "Terraform backend config save path").
		Default(defaultTerraformBackendConfig).
		StringVar(&c.backendConfigPath)

	cmd.Flag("project-config", "Project specific config path").
		Default(defaultProjectConfig).
		StringVar(&c.projectConfigPath)
}

func (c *BackendCommand) run(context *kingpin.ParseContext) error {
	c.dotEnvConfig = dotEnv{
		environment: c.readDotEnv(c.environmentConfigPath),
		project:     c.readDotEnv(c.projectConfigPath),
	}

	c.backendConfig = c.handleDotEnv(&c.dotEnvConfig)

	c.app.AskConfirmOrSkip(c.app.isCi)

	if c.app.createOrPopulateFile(c.backendConfigPath, c.executeTemplate(c.template, c.backendConfig)) {
		c.log.Info("Successfully generated: %s", filepath.Base(c.backendConfigPath))
	} else {
		c.log.ErrorF("I don't really know what exactly should be happen to cause that error ¯\\_(ツ)_/¯ ")
	}

	return nil
}

func (c *BackendCommand) validate(context *kingpin.ParseContext) error {

	c.template = c.parseTemplate(backendTemplate)

	c.app.ValidatePath()

	c.log.ShowOpts("Environment", c.environment)
	if err, isValid := ValidateEnvironment(c.environment); !isValid {
		c.log.ErrorFWithUsage(err)
	}

	modulesAbsPath, isFoundModules := c.findModules(c.app.projectPath)
	if !isFoundModules {
		c.log.ErrorF("Cant find '%s' dir", ModulesDir)
	}
	c.modulesPath = modulesAbsPath

	c.environmentConfigPath = filepath.Join(c.modulesPath, EnvironmentsDir, c.environment, defaultEnvironmentConfig)
	c.log.ShowOpts("Environment config path", c.environmentConfigPath)
	if isExists, _ := ValidateFile(c.environmentConfigPath); !isExists {
		c.log.ErrorF("Environment config '%s' not exists", defaultEnvironmentConfig)
	}

	projectConfigPath, _ := filepath.Abs(c.projectConfigPath)
	c.log.ShowOpts("Project config path", projectConfigPath)
	if isExists, _ := ValidateFile(projectConfigPath); !isExists {
		c.log.ErrorF("Project config '%s' not exists", filepath.Base(projectConfigPath))
	}
	c.projectConfigPath = projectConfigPath

	backendConfigPath, _ := filepath.Abs(c.backendConfigPath)
	c.log.ShowOpts("Backend config path", backendConfigPath)
	if isExists, isWritable := ValidateFile(backendConfigPath); isExists && isWritable {
		c.log.Warning("Backend config '%s' exists and will be overridden", filepath.Base(backendConfigPath))
	} else if isExists && !isWritable {
		c.log.ErrorF("Backend config '%s' exists, but dont have write permissions", filepath.Base(backendConfigPath))
	} else {
		c.log.Info("Backend config '%s' does'nt exists and will be created", filepath.Base(backendConfigPath))
	}
	c.backendConfigPath = backendConfigPath

	return nil
}

func (c *BackendCommand) handleDotEnv(env *dotEnv) *BackendConfig {
	return &BackendConfig{
		Environment:          c.environment,
		Region:               env.environment["REGION"],
		TerraformStateBucket: env.environment["TERRAFORM_STATE_BUCKET"],
		TerraformLockTable:   env.environment["TERRAFORM_LOCK_TABLE"],
		KmsKeyArn:            env.environment["KMS_KEY_ARN"],
		TerraformStateKey:    env.project["TERRAFORM_STATE_KEY"],
	}
}

func (c *BackendCommand) executeTemplate(t *template.Template, config *BackendConfig) string {
	buffer := new(bytes.Buffer)
	err := t.Execute(buffer, config)
	c.log.must(err)

	if p := buffer.String(); p != "" {
		return p
	}

	return ""
}

func (c *BackendCommand) parseTemplate(templateText string) *template.Template {
	t, err := template.New("template").Funcs(template.FuncMap{}).Parse(templateText)
	c.log.must(err)
	return t
}

func (c *BackendCommand) readDotEnv(dotEnvFile string) map[string]string {
	e, err := godotenv.Read(dotEnvFile)
	c.log.must(err)
	return e
}

func (c *BackendCommand) findModules(path string) (modulesPath string, isFound bool) {
	for _, v := range listSearchPaths() {
		searchPath, _ := filepath.Abs(filepath.Join(path, v))
		if c.app.FindFolder(searchPath, ModulesDir) {
			return filepath.Join(searchPath, ModulesDir), true
		}
	}
	return "", false
}
