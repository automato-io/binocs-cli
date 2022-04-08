# binocs-cli

## Release

### 1. Set version in `root.go`

`const BinocsVersion = "v0.4.0"`

### 2. Generate docs and update docs on the website

```shell
$ go run main.go docgen
$ mkdir ~/Code/automato/binocs-website/resources/docs/v0.4.0/
$ cp -a docs/* ~/Code/automato/binocs-website/resources/docs/v0.4.0/
```

Update `routes/web.php` and `config/binocs.php` in web project to include the new version and make it default.

### 3. Execute release via GitHub Actions

```shell
$ git commit -m 'bump version to 0.4.0'
$ git tag -a v0.4.0 -m 'release v0.4.0'
$ git push origin master
$ git push origin v0.4.0
```

Download binaries to `~/Code/automato/binocs-download/public/`

```shell
$ cd ~/Code/automato/binocs-download/
$ cp ~/Downloads/binocs_0.4.0_* public/
$ echo "v0.4.0" > public/VERSION
$ git add .
$ git commit -m 'v0.4.0'
$ git push origin master
$ ./sync
```

## Test cases for valid UpCode regexp pattern

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

## Help pages

### ToC

- [x] [binocs](#binocs)
- [x] [binocs user](#binocs-user)
- [x] [binocs user generate-key](#binocs-user-generate-key)
- [x] [binocs user invalidate-key](#binocs-user-invalidate-key)
- [x] [binocs user update](#binocs-user-update)
- [x] [binocs check](#binocs-check)
- [x] [binocs check add](#binocs-check-add)
- [x] [binocs check delete](#binocs-check-delete)
- [x] [binocs check inspect](#binocs-check-inspect)
- [x] [binocs check list](#binocs-check-list)
- [x] [binocs check update](#binocs-check-update)
- [x] [binocs help](#binocs-help)
- [x] [binocs incident](#binocs-incident)
- [x] [binocs incident inspect](#binocs-incident-inspect)
- [x] [binocs incident list](#binocs-incident-list)
- [x] [binocs incident update](#binocs-incident-update)
- [x] [binocs login](#binocs-login)
- [x] [binocs logout](#binocs-logout)
- [x] [binocs channel](#binocs-channel)
- [x] [binocs channel add](#binocs-channel-add)
- [x] [binocs channel attach](#binocs-channel-attach)
- [x] [binocs channel detach](#binocs-channel-detach)
- [x] [binocs channel list](#binocs-channel-list)
- [x] [binocs channel delete](#binocs-channel-delete)
- [x] [binocs channel update](#binocs-channel-update)
- [x] [binocs channel view](#binocs-channel-view)
- [x] [binocs status](#binocs-status)
- [x] [binocs version](#binocs-version)

### binocs

`binocs --help`

```
Binocs is a CLI-first uptime and performance monitoring tool for websites, applications and APIs.

Binocs servers continuously measure uptime and performance of http or tcp endpoints. 

Get insight into current state of your endpoints and metrics history using this CLI tool. 

Receive notifications about any incidents in real-time.

Usage:
  binocs [command] [flags] [args]

Available Commands:
  user        Manage your binocs user
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

### binocs user

- [ ] implemented

`binocs user --help`

```
Display information about your binocs user.

(name, email, password-***, billing address, timezone)

Usage:
  binocs user [flags]
  binocs user [command] [flags]

Available Commands:
  generate-key      Generate new Access ID and Secret Key
  invalidate-key    Deny future login attempts using this key
  update            Update your binocs user

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output

Use "binocs user [command] --help" for more information about a command.
```

### binocs user generate-key

- [x] implemented

`binocs user generate-key --help`

```
Generate new Access ID and Secret Key.

Usage:
  binocs user generate-key [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

### binocs user invalidate-key

- [x] implemented

`binocs user invalidate-key --help`

```
Deny future login attempts using this key.

Usage:
  binocs user invalidate-key [arg] [flags]

Arg: Access Key

Flags:
  -h, --help    Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

### binocs user update

- [x] implemented

`binocs user update --help`

```
Update any of the following parameters of your user: 
name, timezone

Usage:
  binocs user update [flags]

Flags:
      --name string                        Your name
      --timezone                           Your timezone, defaults to UTC
  -h, --help                               Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

### binocs check

- [x] implemented

`binocs check --help`

```
Manage your checks. A command (one of "add", "delete", "inspect", "list" or "update") is optional.

If neither command nor argument are provided, assume "binocs checks list".

If an argument is provided without any command, assume "binocs checks inspect <arg>".

Usage:
  binocs-cli check [flags]
  binocs-cli check [command]

Aliases:
  check, checks

Available Commands:
  add         add a new endpoint that you want to check
  delete      delete existing check and collected metrics
  inspect     view check status and metrics
  list        list all checks with status and metrics overview
  update      update existing check attributes

Flags:
  -h, --help            help for check
  -p, --period string   display values and charts for specified period (default "day")
  -r, --region string   display values and charts from the specified region only
  -s, --status string   list only "up" or "down" checks, default "all"

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs-cli check [command] --help" for more information about a command.
```

### binocs check add

- [x] implemented

`binocs check add --help`

```
Add a check and start reporting on it. Check identifier is returned upon successful add operation.

This command is interactive and asks user for parameters that were not provided as flags. See the flags overview below.

Usage:
  binocs-cli check add [flags]

Aliases:
  add, create

Flags:
  -n, --name string                        check name
  -u, --url string                         URL to check
  -m, --method string                      HTTP method (GET, HEAD, POST, PUT, DELETE) (default "GET")
  -i, --interval int                       how often binocs checks the URL, in seconds (default 60)
  -t, --target float                       response time that accomodates Apdex=1.0, in seconds with up to 3 decimal places (default 1.2)
  -r, --regions all                        from where in the world we check the provided URL. Choose all or any combination of `us-east-1`, `eu-central-1`, ... (default [all])
      --up_codes 2xx                       what are the good ("UP") HTTP response codes, e.g. 2xx or `200-302`, or `200,301` (default "200-302")
      --up_confirmations_threshold int     how many subsequent Up responses before triggering notifications (default 2)
      --down_confirmations_threshold int   how many subsequent Down responses before triggering notifications (default 2)
  -h, --help                               help for add

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs check delete

- [x] implemented

`binocs check delete --help`

```
Delete existing check and collected metrics.

Usage:
  binocs-cli check delete [flags]

Aliases:
  delete, del, rm

Flags:
  -h, --help   help for delete

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs check inspect

- [x] implemented

`binocs check inspect --help`

```
View check status and metrics.

Usage:
  binocs-cli check inspect [flags]

Aliases:
  inspect, view, show, info

Flags:
  -h, --help            help for inspect
  -p, --period string   display values and charts for specified period (default "day")
  -r, --region string   display values and charts from the specified region only

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs check list

- [x] implemented

`binocs check list --help`

```
List all checks with status and metrics overview.

Usage:
  binocs-cli check list [flags]

Aliases:
  list, ls

Flags:
  -h, --help            help for list
  -p, --period string   display MRT, UPTIME, APDEX values and APDEX chart for specified period (default "day")
  -r, --region string   display MRT, UPTIME, APDEX values and APDEX chart from the specified region only
  -s, --status string   list only "up" or "down" checks, default "all"

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs check update

- [x] implemented

`binocs check update --help`

```
Update existing check attributes.

Usage:
  binocs-cli check update [flags]

Flags:
  -n, --name string                        check name
  -u, --url string                         URL to check
  -m, --method string                      HTTP method (GET, HEAD, POST, PUT, DELETE)
  -i, --interval int                       how often we check the URL, in seconds
  -t, --target float                       response time that accomodates Apdex=1.0, in seconds with up to 3 decimal places
  -r, --regions all                        from where in the world we check the provided URL. Choose all or any combination of `us-east-1`, `eu-central-1`, ...
      --up_codes 2xx                       what are the good ("UP") HTTP response codes, e.g. 2xx or `200-302`, or `200,301`
      --up_confirmations_threshold int     how many subsequent Up responses before triggering notifications
      --down_confirmations_threshold int   how many subsequent Down responses before triggering notifications
  -h, --help                               help for update

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs help

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

### binocs incident

- [ ] implemented

`binocs incident --help`

```
Manage your incidents. A command (one of "inspect", "list" or "update") is optional.

If neither command nor argument are provided, assume `binocs incidents list`.

If an argument is provided without any command, assume `binocs incidents inspect <arg>`.

Usage:
  binocs incident [command] [flags] [arg]

Arg: a 10 characters long incident identifier

Aliases:
  incident, incidents

Available Commands:
  inspect
  list
  update

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output

Use "binocs incident [command] --help" for more information about a command.
```

### binocs incident inspect

- [ ] implemented

`binocs incident inspect --help`

```
View incident information, duration and error codes

Usage:
  binocs incident inspect [arg] [flags]

Aliases:
  inspect, view, show, info

Arg: an incident ID

Flags:
  -h, --help    help for inspect

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

### binocs incident list

- [x] implemented

`binocs incident list --help`

```
List all past incidents.

Usage:
  binocs-cli incident list [flags]

Aliases:
  list, ls

Flags:
  -h, --help   help for list

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs incident update

- [ ] implemented

`binocs incident update --help`

```
Update incident notes

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

### binocs login

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

### binocs logout

- [x] implemented

`binocs logout --help`

```
Logs you out of the binocs user on this machine.

Usage:
  binocs logout [flags]

Flags:
  -h, --help            Display help

Global Flags:
      --config string   Config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         Verbose output
```

### binocs channel

- [x] implemented

`binocs channel --help`

```
Manage notification channels

Usage:
  binocs-cli channel [flags]
  binocs-cli channel [command]

Aliases:
  channel, channels

Available Commands:
  add         add a new notification channel
  inspect     view channel details
  list        list all notification channels
  update      update existing notification channel

Flags:
  -h, --help   help for channel

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output

Use "binocs-cli channel [command] --help" for more information about a command.
```

### binocs channel add

- [x] implemented

`binocs channel add --help`

```
Add a new notification channel

Usage:
  binocs-cli channel add [flags]

Aliases:
  add, create

Flags:
      --alias string    channel alias - how we're gonna refer to it; optional
      --handle string   channel handle - e-mail address for E-mail, Slack URL for Slack
  -h, --help            help for add
  -t, --type string     channel type (E-mail, Slack, Telegram)

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs channel attach

- [x] implemented

`binocs channel attach --help`


```
Attach channel to one or more checks, either for "status", "http-code-change" or both types of notifications

Usage:
  binocs-cli channel attach [flags]

Aliases:
  attach, att

Flags:
  -c, --check string   check identifier, using multiple comma-separated identifiers is supported
  -t, --type string    notification type, "status" or "http-code-change" or both, defaults to "http-code-change,status"
  -h, --help           help for attach

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs channel detach

- [x] implemented

`binocs channel detach --help`

```
Detach channel from one or more checks, either for "status", "http-code-change" or both types of notifications

Usage:
  binocs-cli channel detach [flags]

Flags:
  -c, --check string   check identifier, using multiple comma-separated identifiers is supported
  -t, --type string    notification type, "status" or "http-code-change" or both, defaults to "http-code-change,status"
  -h, --help           help for detach

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs channel list

- [x] implemented

`binocs channel list --help`

```
List all notification channels.

Usage:
  binocs-cli channel list [flags]

Aliases:
  list, ls

Flags:
  -c, --check string   list only notification channels attached to a specific check
  -h, --help           help for list

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs channel delete

- [x] implemented

`binocs channel delete --help`

```
Delete a notification channel.

Usage:
  binocs-cli channel delete [flags]

Aliases:
  delete, del, rm

Flags:
  -h, --help   help for delete

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs channel update

- [x] implemented

`binocs channel update --help`

```
Update existing notification channel.

Usage:
  binocs-cli channel update [flags]

Flags:
      --alias string    channel alias - how we're gonna refer to it; optional
      --handle string   channel handle - e-mail address for E-mail, Slack URL for Slack
  -h, --help            help for update

Global Flags:
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### binocs channel view

- [x] implemented

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

### binocs status

- [ ] implemented

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

### binocs version

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