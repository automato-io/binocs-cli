## binocs channel attach

Attach channel to one or more checks

### Synopsis


Attach channel to one or more checks, either for "status", "http-code-change" or both types of notifications


```
binocs channel attach [flags]
```

### Options

```
  -c, --check string   check identifier, using multiple comma-separated identifiers is supported
  -t, --type string    notification type, "status" or "http-code-change" or both, defaults to "http-code-change,status"
  -a, --all            attach all checks to this channel
  -h, --help           help for attach
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs channel](binocs_channel.md)	 - Manage notification channels

