package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	opts := parseFlags(os.Args[1:])

	command, err := readInput(opts.args, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if command == "" {
		if opts.countOnly {
			fmt.Println(0)
		}
		return
	}

	if opts.namesOnly {
		names, err := commandNames(command)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if opts.countOnly {
			fmt.Println(len(names))
			return
		}
		for _, name := range names {
			fmt.Println(name)
		}
		return
	}

	commands, err := splitCommands(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if opts.countOnly {
		fmt.Println(len(commands))
		return
	}

	for _, cmd := range commands {
		fmt.Println(cmd)
	}
}

type options struct {
	args      []string
	countOnly bool
	namesOnly bool
}

func parseFlags(args []string) options {
	var opts options
	for _, arg := range args {
		switch arg {
		case "-n":
			opts.countOnly = true
		case "-c":
			opts.namesOnly = true
		default:
			opts.args = append(opts.args, arg)
		}
	}
	return opts
}

// readInput は引数またはstdinからコマンド文字列を取得する。
func readInput(args []string, stdin io.Reader) (string, error) {
	if len(args) > 0 {
		return strings.Join(args, " "), nil
	}

	// stdinがターミナルの場合は使い方を表示
	if f, ok := stdin.(*os.File); ok {
		if isTerminal(f) {
			return "", fmt.Errorf("usage: shs <command>\n       echo <command> | shs")
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
