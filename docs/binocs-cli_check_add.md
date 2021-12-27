## binocs-cli check add

Add a new endpoint that you want to check

### Synopsis


Add a check and start reporting on it. Check identifier is returned upon successful add operation.

This command is interactive and asks user for parameters that were not provided as flags. See the flags overview below.


```
binocs-cli check add [flags]
```

### Options

```
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
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs-cli.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs-cli check](binocs-cli_check.md)	 - Manage your checks

