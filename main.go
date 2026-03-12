package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	opts, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if opts.help {
		printUsage()
		return
	}

	command, err := readInput(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if command == "" {
		return
	}

	var results []string
	if opts.namesOnly {
		results, err = commandNames(command)
	} else {
		results, err = splitCommands(command)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if opts.countOnly {
		fmt.Println(len(results))
		return
	}

	for _, r := range results {
		fmt.Println(r)
	}
}

type options struct {
	countOnly bool
	namesOnly bool
	help      bool
}

func parseFlags(args []string) (options, error) {
	var opts options
	for _, arg := range args {
		switch arg {
		case "-n":
			opts.countOnly = true
		case "-c":
			opts.namesOnly = true
		case "-h", "--help":
			opts.help = true
		default:
			return options{}, fmt.Errorf("unexpected argument: %s", arg)
		}
	}
	if opts.countOnly && opts.namesOnly {
		return options{}, fmt.Errorf("-c and -n cannot be used together")
	}
	return opts, nil
}

func printUsage() {
	fmt.Print(`shs - Shell command Splitter

Split chained shell commands (pipes, &&, ||, ;) into individual commands.
Command substitutions $() are also expanded recursively.

Usage:
  echo <command> | shs [options]

Options:
  -c    Extract command names only (e.g., "git" instead of "git log --oneline")
  -n    Print the number of commands instead of the commands themselves
  -h    Show this help

Examples:
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
`)
}

// readInput はstdinからコマンド文字列を取得する。
func readInput(stdin io.Reader) (string, error) {
	// stdinがターミナルの場合は使い方を表示
	if f, ok := stdin.(*os.File); ok {
		if isTerminal(f) {
			return "", fmt.Errorf("no input. Run \"shs -h\" for usage")
		}
	}

	scanner := bufio.NewScanner(stdin)
	var lines []string
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return "", err
	}
	return strings.Join(lines, "\n"), nil
}

func isTerminal(f *os.File) bool {
	stat, err := f.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) != 0
}
