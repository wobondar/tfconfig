package main

const Version = "v0.2.0"

const CiEnvVar = "CI"
const TerraformEnvVar = "TF_ENV"
const ModulesDir = "aws-terraform-modules"
const ConfigFile = "config.tf"
const ConfigModuleName = "config"
const EnvironmentsDir = "environment"
const EnvironmentFile = "environment.tf"

const WarningHeader = `######################################
##   DO NOT EDIT THIS FILE          ##
##   Generated by tfconfig          ##
######################################

`

func main() {
	// everything managed inside cli.go, go away from here mate
	Init()
}
