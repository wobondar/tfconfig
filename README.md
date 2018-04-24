# tfconfig

### Terraform configuration manager

Environment variables:

`CI` - if true or 1, confirmation before any changes will be skipped

`TF_ENV` - environment name (`<environment>` argument will override this value)

```
usage: tfconfig [<flags>] <command> [<args> ...]

Terraform configuration manager

Flags:
  -h, --help       Show context-sensitive help (also try --help-long and --help-man).
  -v, --version    Show application version.
  -c, --ci         CI flag, default 'false', if 'true' that you will not be asked before changes
  -p, --path=PATH  Terraform project path

Commands:
  help [<command>...]
    Show help.

  env <environment>
    Switch project environment
```
