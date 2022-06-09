## binocs channel add

Add a new notifications channel

### Synopsis


Add a new notifications channel.

This command is interactive and asks user for parameters that were not provided as flags.


```
binocs channel add [flags]
```

### Options

```
  -t, --type string      channel type (E-mail, Slack, Telegram)
      --handle string    channel handle - an address for "E-mail" channel type; handles for Slack and Telegram will be obtained programmatically
      --alias string     channel alias (optional)
      --attach strings   checks to attach to this channel (optional); can be either "all", or one or more check identifiers
  -h, --help             help for add
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs channel](binocs_channel.md)	 - Manage notification channels

