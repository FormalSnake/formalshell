This is a minimalistic shell that implements a few basic commands and a simple pipeline.

## Commands

Commands are located in the `cmds` package and are run as functions in the main runtime

- `cd`: Change the current directory.
- `ls`: List the contents of the current directory.
- `exit`: Exit the shell.

## Pipelines

Commands can be chained together using the `&&` operator.
