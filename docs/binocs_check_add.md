## binocs check add

Add a new endpoint that you want to check

### Synopsis


Add a check and start reporting on it. Check identifier is returned upon successful add operation.

This command is interactive and asks user for parameters that were not provided as flags.


```
binocs check add [flags]
```

### Options

```
  -n, --name string                        check name
  -p, --protocol string                    protocol (HTTP, HTTPS or TCP)
  -r, --resource string                    resource to check, a URL in case of HTTP(S), or HOSTNAME:PORT in case of TCP
  -m, --method string                      HTTP(S) method (GET, HEAD, POST, PUT, DELETE)
  -i, --interval int                       how often Binocs checks given resource, in seconds (default 60)
  -t, --target float                       response time that accommodates Apdex=1.0, in seconds with up to 3 decimal places (default 1.2)
      --region strings                     from where in the world Binocs checks given resource; choose one or more from: Australia, Brazil, Germany, Hong Kong, India, Ireland, Japan, Singapore, South Africa, US East, US West
      --up_codes 2xx                       what are the good ("up") HTTP(S) response codes, e.g. 2xx or `200-302`, or `200,301` (default "200-302")
      --up_confirmations_threshold int     how many subsequent "up" responses before triggering notifications (default 2)
      --down_confirmations_threshold int   how many subsequent "down" responses before triggering notifications (default 2)
      --attach strings                     channels to attach to this check (optional); can be either "all", or one or more channel identifiers
  -h, --help                               help for add
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs check](binocs_check.md)	 - Manage checks

