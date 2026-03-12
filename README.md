# shs - Shell command Splitter

A CLI tool that splits chained shell commands (pipes, `&&`, `||`, `;`) into individual commands.

Command substitutions `$()` and backticks are expanded recursively. Subshells `()` are also flattened.

## Installation

### Go

```sh
go install github.com/gurisugi/shs@latest
```

### Homebrew

```sh
brew install gurisugi/tap/shs
```

## Usage

```sh
echo <command> | shs [options]
```

### Options

| Flag | Description |
|------|-------------|
| `-c` | Extract command names only (e.g., `git log --oneline` → `git`) |
| `-n` | Print the number of commands |
| `-h` | Show help |

`-c` and `-n` are mutually exclusive.

### Examples

```sh
$ echo "git log --oneline | wc -l" | shs
git log --oneline
wc -l

$ echo "git log --oneline | wc -l" | shs -c
git
wc

$ echo "git log --oneline | wc -l" | shs -n
2

$ echo 'echo "$(cat file)" && ls' | shs
echo "$()"
cat file
ls
```

## License

[MIT](LICENSE)
