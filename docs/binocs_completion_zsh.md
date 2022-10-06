## binocs completion zsh

Generate the autocompletion script for zsh

### Synopsis

Generate the autocompletion script for the zsh shell.

If shell completion is not already enabled in your environment you will need
to enable it.  You can execute the following once:

	echo "autoload -U compinit; compinit" >> ~/.zshrc

To load completions in your current shell session:

	source <(binocs completion zsh); compdef _binocs binocs

To load completions for every new session, execute once:

#### Linux:

	binocs completion zsh > "${fpath[1]}/_binocs"

#### macOS:

	binocs completion zsh > $(brew --prefix)/share/zsh/site-functions/_binocs

You will need to start a new shell for this setup to take effect.


```
binocs completion zsh [flags]
```

### Options

```
  -h, --help   help for zsh
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs completion](binocs_completion.md)	 - Generate the autocompletion script for the specified shell

