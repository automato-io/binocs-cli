## binocs check add

Add a new endpoint that you want to check

### Synopsis


Add a check and start reporting on it. Check identifier is returned upon successful add operation.

This command is interactive and asks user for parameters that were not provided as flags. See the flags overview below.


```
binocs check add [flags]
```

### Options

```
  -n, --name string                        check name
  -u, --url string                         URL to check
  -m, --method string                      HTTP method (GET, HEAD, POST, PUT, DELETE) (default "GET")
  -i, --interval int                       how often binocs checks the URL, in seconds (default 60)
  -t, --target float                       response time that accomodates Apdex=1.0, in seconds with up to 3 decimal places (default 1.2)
  -r, --regions strings                    from where in the world we check the provided URL. Choose "all" or a combination of values: us-east-1, us-west-1, ap-east-1, ap-southeast-2, eu-central-1, eu-west-1, ap-south-1, ap-northeast-1, ap-southeast-1, sa-east-1, af-south-1
      --up_codes 2xx                       what are the good ("UP") HTTP response codes, e.g. 2xx or `200-302`, or `200,301` (default "200-302")
      --up_confirmations_threshold int     how many subsequent Up responses before triggering notifications (default 2)
      --down_confirmations_threshold int   how many subsequent Down responses before triggering notifications (default 2)
  -h, --help                               help for add
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs check](binocs_check.md)	 - Manage your checks

