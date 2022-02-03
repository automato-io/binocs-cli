## binocs check

Manage your checks

### Synopsis


Manage your checks. A command (one of "add", "delete", "inspect", "list" or "update") is optional.

If neither command nor argument are provided, assume "binocs checks list".
	
If an argument is provided without any command, assume "binocs checks inspect <arg>".


```
binocs check [flags]
```

### Options

```
  -h, --help            help for check
  -p, --period string   display values and charts for specified period (default "day")
  -r, --region string   display values and charts from the specified region only
  -s, --status string   list only "up" or "down" checks, default "all"
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs](binocs.md)	 - Monitoring tool for websites, applications and APIs
* [binocs check add](binocs_check_add.md)	 - Add a new endpoint that you want to check
* [binocs check delete](binocs_check_delete.md)	 - Delete existing check and collected metrics
* [binocs check inspect](binocs_check_inspect.md)	 - View check status and metrics
* [binocs check list](binocs_check_list.md)	 - List all checks with status and metrics overview
* [binocs check update](binocs_check_update.md)	 - Update existing check attributes

