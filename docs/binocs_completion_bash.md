## binocs completion bash

Generate the autocompletion script for bash

### Synopsis

Generate the autocompletion script for the bash shell.

This script depends on the 'bash-completion' package.
If it is not installed already, you can install it via your OS's package manager.

To load completions in your current shell session:

	source <(binocs completion bash)

To load completions for every new session, execute once:

#### Linux:

	binocs completion bash > /etc/bash_completion.d/binocs

#### macOS:

	binocs completion bash > $(brew --prefix)/etc/bash_completion.d/binocs

You will need to start a new shell for this setup to take effect.


```
binocs completion bash
```

### Options

```
  -h, --help   help for bash
```

### Options inherited from parent commands

```
      --config string   config file (default is $HOME/.binocs/config.json)
  -q, --quiet           enable quiet mode (hide spinners and progress bars)
  -v, --verbose         verbose output
```

### SEE ALSO

* [binocs completion](binocs_completion.md)	 - Generate the autocompletion script for the specified shell

