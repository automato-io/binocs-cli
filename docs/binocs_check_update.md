## binocs check update

Update existing check attributes

### Synopsis


Update existing check attributes.


```
binocs check update [flags]
```

### Options

```
  -n, --name string                        check name
  -u, --url string                         URL to check
  -m, --method string                      HTTP method (GET, HEAD, POST, PUT, DELETE)
  -i, --interval int                       how often we check the URL, in seconds
  -t, --target float                       response time that accomodates Apdex=1.0, in seconds with up to 3 decimal places
  -r, --regions strings                    from where in the world we check the provided URL. Choose "all" or a combination of values: us-west-1, ap-east-1, ap-south-1, ap-southeast-1, eu-west-1, af-south-1, us-east-1, ap-northeast-1, ap-southeast-2, eu-central-1, sa-east-1
      --up_codes 2xx                       what are the good ("UP") HTTP response codes, e.g. 2xx or `200-302`, or `200,301`
      --up_confirmations_threshold int     how many subsequent Up responses before triggering notifications
      --down_confirmations_threshold int   how many subsequent Down responses before triggering notifications
  -h, --help                               help for update
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs check](binocs_check.md)	 - Manage your checks

