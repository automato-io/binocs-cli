## binocs completion fish

Generate the autocompletion script for fish

### Synopsis

Generate the autocompletion script for the fish shell.

To load completions in your current shell session:

	binocs completion fish | source

To load completions for every new session, execute once:

	binocs completion fish > ~/.config/fish/completions/binocs.fish

You will need to start a new shell for this setup to take effect.


```
binocs completion fish [flags]
```

### Options

```
  -h, --help   help for fish
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs completion](binocs_completion.md)	 - Generate the autocompletion script for the specified shell

