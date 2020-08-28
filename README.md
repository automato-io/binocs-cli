# binocs-cli

# help pages

## ToC

- [binocs](#binocs)
- [binocs account](#binocs-account)
- [binocs account update](#binocs-account-update)
- [binocs check](#binocs-check)
- [binocs check add](#binocs-check-add)
- [binocs check delete](#binocs-check-delete)
- [binocs check inspect](#binocs-check-inspect)
- [binocs check list](#binocs-check-list)
- [binocs check update](#binocs-check-update)
- [binocs help](#binocs-help)
- [binocs incident](#binocs-incident)
- [binocs incident view](#binocs-incident-view)
- [binocs login](#binocs-login)
- [binocs channel](#binocs-channel)
- [binocs channel add](#binocs-channel-add)
- [binocs channel associate](#binocs-channel-associate)
- [binocs channel disassociate](#binocs-channel-disassociate)
- [binocs channel list](#binocs-channel-list)
- [binocs channel remove](#binocs-channel-remove)
- [binocs channel update](#binocs-channel-update)
- [binocs status](#binocs-status)
- [binocs version](#binocs-version)

### binocs

`binocs --help`

```
binocs is a devops-oriented monitoring tool for websites, applications and APIs.

binocs continuously measures uptime and performance of http or tcp endpoints
and provides insight into current state and metrics history.
Get notified via Slack, Telegram and SMS.

Usage:
  binocs [command]

Available Commands:
  account     Manage your binocs account
  check       Manage your checks
  help        Help about any command
  incident    Manage incidents
  login       Login to binocs
  channel     Manage your notification channels
  status      Display binocs service status info
  version     Print the version number of binocs

Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -h, --help            help for binocs-cli
  -t, --toggle          Help message for toggle
  -v, --verbose         verbose output

Use "binocs [command] --help" for more information about a command.
```

### binocs account

`binocs account --help`

```
Display information about your binocs user account.

Usage:
  binocs account [flags]

Available Commands:
  update     Manage your binocs account

Flags:
  -h, --help   help for account

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs account [command] --help" for more information about a command.
```

### binocs account update

`binocs account update --help`

```
Usage:
  binocs account update [flags]

Flags:
  -e, --example string                        Example
  -h, --help                               help for add

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs check

`binocs check --help`

```
Manage your checks.

Usage:
  binocs check [flags]
  binocs check [command]

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

Use "binocs check [command] --help" for more information about a command.
```

## binocs check add

`binocs check add --help`

### binocs check delete

`binocs check delete --help`

## binocs check inspect

`binocs check inspect --help`

## binocs check list

`binocs check list --help`

## binocs check update

`binocs check update --help`

## binocs help

`binocs help --help`

```
Help provides help for any command in the application.
Simply type binocs help [path to command] for full details.

Usage:
  binocs help [command] [flags]

Flags:
  -h, --help   help for help

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs incident

`binocs incident --help`

```
...

Usage:
  binocs incident [flags]
  binocs incident [command]

Aliases:
  incident, incidents

Available Commands:
  view

Flags:
  -h, --help   help for incident

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs incident [command] --help" for more information about a command.
```

## binocs incident view

`binocs incident view --help`

## binocs login

`binocs login --help`

```
Login to binocs using your Access ID and Secret Key. 

Usage:
  binocs login [flags]

Aliases:
  login, auth

Flags:
  -h, --help   help for login

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs channel

`binocs channel --help`

## binocs channel add

`binocs channel add --help`

## binocs channel associate

`binocs channel associate --help`

## binocs channel disassociate

`binocs channel disassociate --help`

## binocs channel list

`binocs channel list --help`

## binocs channel remove

`binocs channel remove --help`

## binocs channel update

`binocs channel update --help`

## binocs status

`binocs status --help`

## binocs version

`binocs version --help`

```
All software has versions. This is binocs's

Usage:
  binocs version [flags]

Flags:
  -h, --help   help for version

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```