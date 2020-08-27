# CLI Documentation

`go run main.go --help`

```
binocs is a devops-oriented monitoring tool for websites, applications and APIs.

binocs continuously measures uptime and performance of http or tcp endpoints
and provides insight into current state and metrics history.
Get notified via Slack, Telegram and SMS.

Usage:
  binocs-cli [command]

Available Commands:
  account     Manage your binocs account
  check       Manage your checks
  help        Help about any command
  incident    Manage incidents
  login       Login to binocs
  version     Print the version number of binocs

Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -h, --help            help for binocs-cli
  -t, --toggle          Help message for toggle
  -v, --verbose         verbose output

Use "binocs-cli [command] --help" for more information about a command.
```

`go run main.go account --help`

```
...

Usage:
  binocs-cli account [flags]

Flags:
  -h, --help   help for account

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

`go run main.go check --help`

```
...

Usage:
  binocs-cli check [flags]
  binocs-cli check [command]

Aliases:
  check, checks

Available Commands:
  add
  delete
  inspect
  list
  update

Flags:
  -h, --help            help for check
  -r, --region string   Display MRT, UPTIME and APDEX from the specified region only
  -s, --status string   List only "up" or "down" checks, default "all"

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs-cli check [command] --help" for more information about a command.
```

`go run main.go check add --help`
`go run main.go check delete --help`
`go run main.go check inspect --help`
`go run main.go check list --help`
`go run main.go check update --help`

`go run main.go help --help`

```
Help provides help for any command in the application.
Simply type binocs-cli help [path to command] for full details.

Usage:
  binocs-cli help [command] [flags]

Flags:
  -h, --help   help for help

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

`go run main.go incident --help`

```
...

Usage:
  binocs-cli incident [flags]
  binocs-cli incident [command]

Aliases:
  incident, incidents

Available Commands:
  inspect

Flags:
  -h, --help   help for incident

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs-cli incident [command] --help" for more information about a command.
```

`go run main.go incident inspect --help`

`go run main.go login --help`

```
Login to binocs using your Access ID and Secret Key. 

Usage:
  binocs-cli login [flags]

Aliases:
  login, auth

Flags:
  -h, --help   help for login

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

`go run main.go version --help`

```
All software has versions. This is binocs's

Usage:
  binocs-cli version [flags]

Flags:
  -h, --help   help for version

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```