## binocs checks

List all checks with status and metrics overview

### Synopsis


List all checks with status and metrics overview.


```
binocs checks [flags]
```

### Options

```
  -h, --help            help for checks
  -p, --period string   display MRT, UPTIME, APDEX values and APDEX chart for specified period (default "day")
  -r, --region string   display MRT, UPTIME, APDEX values and APDEX chart from the specified region only
  -s, --status string   list only "up" or "dow" checks, default "all"
      --watch           run in cell view and refresh binocs output every 5 seconds
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs](binocs.md)	 - Monitoring tool for websites, applications and APIs

