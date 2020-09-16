# binocs-cli

# test cases for valid UpCode regexp pattern

```
404
2xx
30x
200-301
200-301,404
404,200-301
200-202,300-302
200-202,300-302,404
200-202,404,300-302
200-202,300-302,403-404
---
200-301-404
200-2xx,404,300-302
099
4044
5xxx
20
2x
2
x
---
200-101
```

# help pages

## ToC

- [x] [binocs](#binocs)
- [x] [binocs account](#binocs-account)
- [x] [binocs account generate-key](#binocs-account-generate-key)
- [x] [binocs account invalidate-key](#binocs-account-invalidate-key)
- [x] [binocs account update](#binocs-account-update)
- [x] [binocs check](#binocs-check)
- [x] [binocs check add](#binocs-check-add)
- [x] [binocs check delete](#binocs-check-delete)
- [x] [binocs check inspect](#binocs-check-inspect)
- [x] [binocs check list](#binocs-check-list)
- [x] [binocs check update](#binocs-check-update)
- [x] [binocs help](#binocs-help)
- [x] [binocs incident](#binocs-incident)
- [x] [binocs incident view](#binocs-incident-view)
- [x] [binocs incident update](#binocs-incident-update)
- [x] [binocs login](#binocs-login)
- [x] [binocs logout](#binocs-logout)
- [x] [binocs channel](#binocs-channel)
- [x] [binocs channel add](#binocs-channel-add)
- [ ] [binocs channel associate](#binocs-channel-associate)
- [ ] [binocs channel disassociate](#binocs-channel-disassociate)
- [x] [binocs channel list](#binocs-channel-list)
- [x] [binocs channel remove](#binocs-channel-remove)
- [x] [binocs channel update](#binocs-channel-update)
- [x] [binocs channel view](#binocs-channel-view)
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
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -h, --help            Display help
  -v, --verbose         Verbose output

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
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output

Use "binocs account [command] --help" for more information about a command.
```

### binocs account generate-key

`binocs account generate-key --help`

```
Generate new Access ID and Secret Key.

Usage:
  binocs account generate-key [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
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
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
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
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs check

- [x] implemented

`binocs check --help`

```
Manage your checks. A command (one of "add", "delete", "inspect", "list" or "update") is optional.

If neither command nor argument are provided, assume `binocs checks list`.

If an argument is provided without any command, assume `binocs checks inspect <arg>`.

Usage:
  binocs check [command] [flags] [arg]

Aliases:
  check, checks

Arg: a 7 characters long check identifier

Available Commands:
  add
  delete
  inspect
  list
  update

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output

Use "binocs check [command] --help" for more information about a command.
```

## binocs check add

- [x] implemented

`binocs check add --help`

```
Add a check and start reporting on it. Check identifier is returned upon successful add operation.

This command is interactive and asks user for parameters that were not provided as flags. See the flags overview below.

Usage:
  binocs check add [flags]

Flags:
  -n, --name string                        Check alias
  -u, --url string                         URL to check
  -m, --method string                      HTTP method (GET, HEAD, POST, PUT, DELETE)
  -i, --interval int                       How often we check the URL, in seconds (default 30)
  -t, --target float                       Response time in miliseconds for Apdex = 1.0 (default 0.7)
  -r, --regions all                        From where we check the URL, choose all or any combination of `us-east-1`, `eu-central-1`, ... (default [all])
      --up_codes string                    What are the Up HTTP response codes, e.g. 2xx or `200-302`, or `200,301` (default "200-302")
      --up_confirmations_threshold int     How many subsequent Up responses before triggering notifications (default 2)
      --down_confirmations_threshold int   How many subsequent Down responses before triggering notifications (default 2)
  -h, --help                               Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

### binocs check delete

- [x] implemented

`binocs check delete --help`

```
Delete a check

Usage:
  binocs check delete [flags] [arg]

Arg: a 7 characters long check identifier

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs check inspect

- [ ] implemented

`binocs check inspect --help`

```
View detailed info about check's status and history.

Usage:
  binocs check inspect [flags] [arg]

Aliases:
  inspect, view, show

Arg: a 7 characters long check identifier

Flags:
  -h, --help            Display help
  -r, --region string   Display MRT, UPTIME and APDEX from the specified region only
  -s, --status string   List only "up" or "down" checks, default "all"

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs check list

- [x] implemented

`binocs check list --help`

```
List all checks with status and basic metrics info

Usage:
  binocs check list [flags]

Aliases:
  list, ls

Flags:
  -h, --help              Display help
  -r, --region string     Display MRT, UPTIME and APDEX from the specified region only
  -s, --status string     List only "up" or "down" checks, default "all"

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs check update

- [x] implemented

`binocs check update --help`

```
Update a check and continue reporting on it

Usage:
  binocs check update [flags] [arg]

Flags:
  -n, --name string                        Check alias
  -u, --url string                         URL to check
  -m, --method string                      HTTP method (GET, HEAD, POST, PUT, DELETE)
  -i, --interval int                       How often we check the URL, in seconds (default 30)
  -t, --target float                       Response time in miliseconds for Apdex = 1.0 (default 0.7)
  -r, --regions all                        From where we check the URL, choose all or any combination of `us-east-1`, `eu-central-1`, ... (default [all])
      --up_codes 2xx                       What are the Up HTTP response codes, e.g. 2xx or `200-302`, or `200,301` (default "200-302")
      --up_confirmations_threshold int     How many subsequent Up responses before triggering notifications (default 2)
      --down_confirmations_threshold int   How many subsequent Down responses before triggering notifications (default 2)
  -h, --help                               Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs help

- [x] implemented

`binocs help --help`

```
Help provides help for any command in the application.

Usage:
  binocs help [command] [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
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
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output

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
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
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
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs login

- [x] implemented

`binocs login --help`

```
Login to binocs using your Access ID and Secret Key. 

Usage:
  binocs login [flags]

Aliases:
  login, auth

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs logout

`binocs logout --help`

```
Logs you out of the binocs account on this machine.

Usage:
  binocs logout [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs channel

`binocs channel --help`

```
Manage your notification channels. Use a subcommand, or inspect a channel, if a valid channel _id_ is given as the argument.

Usage:
  binocs channel [flags] [arg]
  binocs channel [command] [flags]

Aliases:
  channel, channels

Arg: a channel ID

Available Commands:
  add
  associate
  disassociate
  list
  remove
  update
  view

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output

Use "binocs channel [command] --help" for more information about a command.
```

## binocs channel add

`binocs channel add --help`

```
Add a channel. Remember to associate channel with your checks using:
  binocs channel associate

Usage:
  binocs channel add [flags]

Flags:
      --type string      Type, one of: sms, slack, telegram, email
      --alias string     Optional name of the channel
      --handle string    Depending on the value of the --type flag
            - email - one or more comma-separated e-mail addresses, each address' first use requires e-mail opt-in
            - sms - a phone number
            - slack - a Slack handle URL
            - telegram - @todo
  -h, --help             Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs channel associate

`binocs channel associate --help`

## binocs channel disassociate

`binocs channel disassociate --help`

## binocs channel list

`binocs channel list --help`

```
List all channels with stats

Usage:
  binocs channels list [flags]

Aliases:
  list, ls

Flags:
      --check string    List only channels associated with this check
  -h, --help            Display help
  -r, --region string   Display MRT, UPTIME and APDEX from the specified region only
  -s, --status string   List only "up" or "down" checks, default "all"

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs channel remove

`binocs channel remove --help`

```
Delete a channel. This also diassociates this channel from all checks. You will stop receiving alerts via this channel once you delete it using this command.

Usage:
  binocs channel delete [flags] [arg]

Arg: a channel ID

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs channel update

`binocs channel update --help`

```
Update a channel

Usage:
  binocs channel update [flags] [arg]

Flags:
      --type string      Type, one of: sms, slack, telegram, email
      --alias string     Optional name of the channel
      --handle string    Depending on the value of the --type flag
            - email - one or more comma-separated e-mail addresses, each address' first use requires e-mail opt-in
            - sms - a phone number
            - slack - a Slack handle URL
            - telegram - @todo
  -h, --help             Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs channel view

`binocs channel view --help`

```
View detailed info about channel

Usage:
  binocs channel view [flags] [arg]

Aliases:
  view, inspect, show

Arg: a channel ID

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs status

`binocs status --help`

```
Display binocs service status info

Usage:
  binocs status [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

## binocs version

- [x] implemented

`binocs version --help`

```
Prints binocs client version

Usage:
  binocs version [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```