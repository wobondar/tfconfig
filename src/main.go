package main

const Version = "v0.4.2"

const CiEnvVar = "CI"
const TerraformLocalEnvVar = "TF_LOCAL"
const TerraformEnvVar = "TF_ENV"
const ModulesDir = "aws-terraform-modules"
const ConfigFile = "config.tf"
const ConfigModuleName = "config"
const EnvironmentsDir = "environment"
const EnvironmentFile = "environment.tf"

// Global configuration that includes environment (dev, staging) configuration stored there
const defaultEnvironmentConfig = "environment.env"

// Project specific configuration
const defaultProjectConfig = "terraform.env"

const WarningHeader = `######################################
##   DO NOT EDIT THIS FILE          ##
##   Generated by tfconfig          ##
######################################

`

func main() {
	// everything managed inside cli.go, go away from here mate
	Init()
}
