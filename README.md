# binocs-cli

# help pages

## ToC

- [x] [binocs](#binocs)
- [x] [binocs account](#binocs-account)
- [x] [binocs account generate-key](#binocs-account-generate-key)
- [x] [binocs account invalidate-key](#binocs-account-invalidate-key)
- [x] [binocs account update](#binocs-account-update)
- [x] [binocs check](#binocs-check)
- [ ] [binocs check add](#binocs-check-add)
- [ ] [binocs check delete](#binocs-check-delete)
- [ ] [binocs check inspect](#binocs-check-inspect)
- [ ] [binocs check list](#binocs-check-list)
- [ ] [binocs check update](#binocs-check-update)
- [x] [binocs help](#binocs-help)
- [x] [binocs incident](#binocs-incident)
- [x] [binocs incident view](#binocs-incident-view)
- [x] [binocs incident update](#binocs-incident-update)
- [x] [binocs login](#binocs-login)
- [x] [binocs logout](#binocs-logout)
- [ ] [binocs channel](#binocs-channel)
- [ ] [binocs channel add](#binocs-channel-add)
- [ ] [binocs channel associate](#binocs-channel-associate)
- [ ] [binocs channel disassociate](#binocs-channel-disassociate)
- [ ] [binocs channel list](#binocs-channel-list)
- [ ] [binocs channel remove](#binocs-channel-remove)
- [ ] [binocs channel update](#binocs-channel-update)
- [x] [binocs status](#binocs-status)
- [x] [binocs version](#binocs-version)

### binocs

`binocs --help`

```
binocs is a devops-oriented monitoring tool for websites, applications and APIs.

binocs continuously measures uptime and performance of http or tcp endpoints
and provides insight into current state and metrics history.
Get notified via Slack, Telegram and SMS.

Usage:
  binocs [command] [flags] [args]

Available Commands:
  account     Manage your binocs account
  check       Manage your checks
  help        Help about any command in the application
  incident    Manage your incidents
  login       Login to binocs
  logout      Log out of binocs
  channel     Manage your notification channels
  status      Display binocs service status info
  version     Print binocs client version

Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -h, --help            help for binocs-cli
  -v, --verbose         verbose output

Use "binocs [command] --help" for more information about a command.
```

### binocs account

`binocs account --help`

```
Display information about your binocs account.

(name, email, password-***, billing address, timezone)

Usage:
  binocs account [flags]
  binocs account [command] [flags]

Available Commands:
  generate-key      Generate new Access ID and Secret Key
  invalidate-key    Deny future login attempts using this key
  update            Update your binocs account

Flags:
  -h, --help   help for account

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs account [command] --help" for more information about a command.
```

### binocs account generate-key

`binocs account generate-key --help`

```
Generate new Access ID and Secret Key.

Usage:
  binocs account generate-key [flags]

Flags:
  -h, --help    Display help

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs account invalidate-key

`binocs account invalidate-key --help`

```
Deny future login attempts using this key.

Usage:
  binocs account invalidate-key [arg] [flags]

Arg: Access ID

Flags:
      --id      The Access ID to invalidate
  -h, --help    Display help

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs account update

`binocs account update --help`

```
Update any of the following parameters of your account: 
email, password, name, billing-address, timezone

Usage:
  binocs account update [flags]

Flags:
      --email string                       Email address, also used as the username
      --password string                    Account password (min. 8 chars)
      --name string                        Account name (Optional)
      --billing-address                    We use it on the invoices only
      --timezone                           Display all times in this timezone, defaults to UTC (London)
  -h, --help                               Display help

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs check

`binocs check --help`

```
Manage your checks. Use a subcommand, or inspect a check, if a valid check _id_ is given as the argument.

Usage:
  binocs check [flags] [arg]
  binocs check [command] [flags]

Aliases:
  check, checks

Arg: a check ID

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

Usage:
  binocs help [command] [flags]

Flags:
  -h, --help   help for help

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs incident

`binocs incident --help`

```
Manage your incidents

Usage:
  binocs incident [flags] [arg]
  binocs incident [command] [flags]

Arg: an incident ID

Aliases:
  incident, incidents

Available Commands:
  view
  update

Flags:
  -h, --help   help for incident

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs incident [command] --help" for more information about a command.
```

## binocs incident view

`binocs incident view --help`

```
View all info about any incident recorded by binocs.

Usage:
  binocs incident view [arg] [flags]

Arg: an incident ID

Flags:
  -h, --help    Display help

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs incident update

`binocs incident update --help`

```
Update incident notes.

Usage:
  binocs incident update [arg] [flags]

Arg: an incident ID

Flags:
  -n, --note    Set incident note to this value
  -h, --help    Display help

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

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

## binocs logout

`binocs logout --help`

```
Logs you out of the binocs account on this machine.

Usage:
  binocs logout [flags]

Flags:
  -h, --help   help for logout

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

```
Display binocs service status info

Usage:
  binocs status [flags]

Flags:
  -h, --help   help for version

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

## binocs version

`binocs version --help`

```
Prints binocs client version

Usage:
  binocs version [flags]

Flags:
  -h, --help   help for version

Global Flags:
  -c, --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```