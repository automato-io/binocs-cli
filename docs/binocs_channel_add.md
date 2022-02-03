## binocs channel add

Add a new notification channel

### Synopsis


Add a new notification channel


```
binocs channel add [flags]
```

### Options

```
  -t, --type string     channel type (E-mail, Slack, Telegram)
      --handle string   channel handle - e-mail address for E-mail, Slack URL for Slack
      --alias string    channel alias - how we're gonna refer to it; optional
  -h, --help            help for add
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs channel](binocs_channel.md)	 - Manage notification channels

