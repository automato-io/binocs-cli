## binocs completion powershell

Generate the autocompletion script for powershell

### Synopsis

Generate the autocompletion script for powershell.

To load completions in your current shell session:

	binocs completion powershell | Out-String | Invoke-Expression

To load completions for every new session, add the output of the above command
to your powershell profile.


```
binocs completion powershell [flags]
```

### Options

```
  -h, --help              help for powershell
      --no-descriptions   disable completion descriptions
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs completion](binocs_completion.md)	 - Generate the autocompletion script for the specified shell

