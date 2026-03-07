package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
)

func main() {
	args, countOnly := parseFlags(os.Args[1:])

	command, err := readInput(args, os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if command == "" {
		if countOnly {
			fmt.Println(0)
		}
		return
	}

	commands, err := splitCommands(command)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if countOnly {
		fmt.Println(len(commands))
		return
	}

	for _, cmd := range commands {
		fmt.Println(cmd)
	}
}

// parseFlags はフラグを解析し、残りの引数とオプションを返す。
func parseFlags(args []string) (remaining []string, countOnly bool) {
	for _, arg := range args {
		switch arg {
		case "-n":
			countOnly = true
		default:
			remaining = append(remaining, arg)
		}
	}
	return
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
