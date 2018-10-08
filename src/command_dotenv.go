package main

import (
	"bytes"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ssm"
	"github.com/joho/godotenv"
	"gopkg.in/alecthomas/kingpin.v2"
	"strings"
	"text/template"
)

const (
	// defaultTemplate is the default template used to determine what the SSM
	// parameter name is for an environment variable.
	defaultTemplate = `{{ if hasPrefix .Value "ssm://" }}{{ trimPrefix .Value "ssm://" }}{{ end }}`

	// defaultBatchSize is the default number of parameters to fetch at once.
	// The SSM API limits this to a maximum of 10 at the time of writing.
	defaultBatchSize = 10

	defaultDotEnvFilePrefix = ".env."
)

// templateFuncs are helper functions provided to the template.
var templateFuncs = template.FuncMap{
	"contains":   strings.Contains,
	"hasPrefix":  strings.HasPrefix,
	"hasSuffix":  strings.HasSuffix,
	"trimPrefix": strings.TrimPrefix,
	"trimSuffix": strings.TrimSuffix,
	"trimSpace":  strings.TrimSpace,
	"trimLeft":   strings.TrimLeft,
	"trimRight":  strings.TrimRight,
	"trim":       strings.Trim,
	"title":      strings.Title,
	"toTitle":    strings.ToTitle,
	"toLower":    strings.ToLower,
	"toUpper":    strings.ToUpper,
}

type ssmVar struct {
	envVar    string
	parameter string
}

type ssmClient interface {
	GetParameters(*ssm.GetParametersInput) (*ssm.GetParametersOutput, error)
}

type DotEnvCommand struct {
	app              *App
	log              *Log
	environment      string
	dotEnvFilePrefix string
	dotEnvFileSource string
	dotEnvFileOut    string
	dotEnvMap        map[string]string
	decrypt          bool
	exposeVars       bool
	exportVars       bool
	template         *template.Template
	ssm              ssmClient
	batchSize        int
}

func ConfigureDotEnvCommand(a *App) {
	c := &DotEnvCommand{
		app:              a,
		log:              a.log,
		dotEnvFilePrefix: defaultDotEnvFilePrefix,
		batchSize:        defaultBatchSize,
		exposeVars:       true,
	}

	cmd := a.cli.Command("dotenv", "Generate .env file or expose configuration into env vars from Parameter Store").
		PreAction(c.validate).
		Action(c.run)

	cmd.Arg("environment", "Environment name").
		Required().
		StringVar(&c.environment)

	cmd.Arg("dotEnvFile", "dotEnv file that the configuration will be saved instead of exposing into env vars").
		StringVar(&c.dotEnvFileOut)

	cmd.Flag("decrypt", "Will attempt to decrypt the parameter, default: true. use --no-decrypt to disable it").
		Default("true").
		Short('d').
		BoolVar(&c.decrypt)

	cmd.Flag("export", "Prints vars prepared for export to env via eval like 'export VAR_NAME=var_value\\n'").
		Default("false").
		Short('e').
		BoolVar(&c.exportVars)
}

func (c *DotEnvCommand) initSsmClient() {
	awsConfig := &aws.Config{
		LogLevel: aws.LogLevel(aws.LogOff),
	}

	if c.log.awsDebug {
		awsConfig.LogLevel = aws.LogLevel(aws.LogDebugWithHTTPBody)
	}
	awsSession, err := session.NewSession(awsConfig)
	c.log.must(err)
	c.ssm = ssm.New(awsSession)
}

func (c *DotEnvCommand) run(context *kingpin.ParseContext) error {

	c.template = c.parseTemplate(defaultTemplate)

	c.dotEnvMap = c.readDotEnv(GetFullPath(c.app.projectPath, c.dotEnvFileSource))

	c.initSsmClient()
	c.processDotEnv()

	c.handleDotEnv()

	return nil
}

func (c *DotEnvCommand) validate(context *kingpin.ParseContext) error {
	if c.dotEnvFileOut == "" {
		c.log.Quite()
	}

	c.app.ValidatePath()

	c.log.ShowOpts("Environment", c.environment)

	if err, isValid := ValidateEnvironment(c.environment); !isValid {
		c.log.ErrorFWithUsage(err)
	}
	c.dotEnvFileSource = c.dotEnvFilePrefix + c.environment

	c.log.ShowOpts("Source dotEnv file", c.dotEnvFileSource)
	if isExists, _ := ValidateFile(GetFullPath(c.app.projectPath, c.dotEnvFileSource)); !isExists {
		c.log.ErrorFWithUsage("dotEnv file: '%s' does'nt exists", c.dotEnvFileSource)
	}

	if c.dotEnvFileOut != "" {
		c.log.ShowOpts("Destination dotEnv file", c.dotEnvFileOut)
		if isExists, isWritable := ValidateFile(GetFullPath(c.app.projectPath, c.dotEnvFileOut)); isExists && !isWritable {
			c.log.ErrorF("dotEnv file: '%s' exists, but does'nt have write permissions", c.dotEnvFileOut)
		} else if isExists && isWritable {
			c.log.Warning("dotEnv file '%s' exists and will be overridden", c.dotEnvFileOut)
		}

		if strings.EqualFold(c.dotEnvFileSource, c.dotEnvFileOut) {
			c.log.ErrorF("Source dotEnv file '%s' and destination dotEnv file '%s' must be different", c.dotEnvFileSource, c.dotEnvFileOut)
		}
		c.exposeVars = false
	}

	return nil
}

func (c *DotEnvCommand) handleDotEnv() {
	switch c.exposeVars {
	case true:
		switch c.exportVars {
		case true:
			c.printExportEnvVars()
		case false:
			c.printEnvVars()
		}
	case false:
		c.app.AskConfirmOrSkip(c.app.isCi)
		c.writeDotEnv(c.dotEnvFileOut, c.dotEnvMap)
	}
}

func (c *DotEnvCommand) printEnvVars() {
	for k, v := range c.dotEnvMap {
		c.log.Printf("%s=%s ", k, v)
	}
}

func (c *DotEnvCommand) printExportEnvVars() {
	for k, v := range c.dotEnvMap {
		c.log.Printf("export %s=\"%s\"\n", k, v)
	}
}

func (c *DotEnvCommand) processDotEnv() error {
	var ssmVars []ssmVar

	uniqNames := make(map[string]bool)
	for k, v := range c.dotEnvMap {
		parameter, err := c.parameter(k, v)
		c.log.must(err)

		if parameter != nil {
			uniqNames[*parameter] = true
			ssmVars = append(ssmVars, ssmVar{k, *parameter})
		}
	}

	c.log.Debug("uniqNames = %v", len(uniqNames))

	if len(uniqNames) == 0 {
		// Nothing to do, no SSM parameters.
		return nil
	}

	names := make([]string, len(uniqNames))
	i := 0
	for k := range uniqNames {
		names[i] = k
		i++
	}

	c.log.Debug("names = %v", len(names))

	for i := 0; i < len(names); i += c.batchSize {
		j := i + c.batchSize
		if j > len(names) {
			j = len(names)
		}

		c.log.Debug("Batch [%v-%v], names: %v, ssmVars: %v", i, j, len(names[i:j]), len(ssmVars[i:j]))

		values, err := c.getParameters(names[i:j], c.decrypt, ssmVars)
		c.log.Debug("Batch [%v-%v], values: %v", i, j, len(values))
		if err != nil {
			return err
		}

		for _, v := range ssmVars {
			val, ok := values[v.parameter]
			if ok {
				c.dotEnvMap[v.envVar] = val
			}
		}
	}

	return nil
}

func (c *DotEnvCommand) getParameters(names []string, decrypt bool, ssmVars []ssmVar) (map[string]string, error) {
	values := make(map[string]string)

	input := &ssm.GetParametersInput{
		WithDecryption: aws.Bool(decrypt),
	}

	for _, n := range names {
		input.Names = append(input.Names, aws.String(n))
	}

	c.log.Debug("REQ: Batch [%v], input.Name: %v", len(ssmVars), len(input.Names))

	resp, err := c.ssm.GetParameters(input)
	c.log.Debug("RESP: Batch [%v], resp.Parameters: %v", len(ssmVars), len(resp.Parameters))

	c.log.must(err)

	for _, v := range ssmVars {
		for _, p := range resp.Parameters {
			if strings.EqualFold(v.parameter, *p.Name) {
				values[v.parameter] = *p.Value
			}
		}
		for _, p := range resp.InvalidParameters {
			if strings.EqualFold(v.parameter, *p) {
				values[v.parameter] = "VALUE_NOT_EXISTS"
				c.log.Warning("Value for parameter: %s not exists in AWS Parameter Store. Environment variable: %s", v.parameter, v.envVar)
			}
		}
	}

	return values, nil
}

func (c *DotEnvCommand) parameter(k, v string) (*string, error) {
	b := new(bytes.Buffer)
	if err := c.template.Execute(b, struct{ Name, Value string }{k, v}); err != nil {
		return nil, err
	}

	if p := b.String(); p != "" {
		return &p, nil
	}

	return nil, nil
}

func (c *DotEnvCommand) parseTemplate(templateText string) *template.Template {
	t, err := template.New("template").Funcs(templateFuncs).Parse(templateText)
	c.log.must(err)
	return t
}

func (c *DotEnvCommand) writeDotEnv(dotEnvFile string, dotEnvMap map[string]string) {
	err := godotenv.Write(dotEnvMap, GetFullPath(c.app.projectPath, dotEnvFile))
	c.log.must(err)
	c.log.Info("Successful.")
}

func (c *DotEnvCommand) readDotEnv(dotEnvFile string) map[string]string {
	e, err := godotenv.Read(dotEnvFile)
	c.log.must(err)
	return e
}
