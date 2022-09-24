## binocs check update

Update attributes of an existing check

### Synopsis


Update attributes of an existing check.

This command is interactive and asks user for parameters that were not provided as flags.


```
binocs check update [flags]
```

### Options

```
  -n, --name string                        check name
  -m, --method string                      HTTP(S) method (GET, HEAD, POST, PUT, DELETE)
  -i, --interval int                       how often Binocs checks given resource, in seconds
  -t, --target float                       response time that accommodates Apdex=1.0, in seconds with up to 3 decimal places
  -r, --region strings                     from where in the world Binocs checks given resource; choose one or more from: Australia, Brazil, Germany, Hong Kong, India, Ireland, Japan, Singapore, South Africa, US East, US West
      --up_codes 2xx                       what are the good ("up") HTTP(S) response codes, e.g. 2xx or `200-302`, or `200,301`
      --up_confirmations_threshold int     how many subsequent "up" responses before triggering notifications
      --down_confirmations_threshold int   how many subsequent "down" responses before triggering notifications
      --attach strings                     channels to attach to this check (optional); can be either "all", or one or more channel identifiers
  -h, --help                               help for update
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs check](binocs_check.md)	 - Manage checks

