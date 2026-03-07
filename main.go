package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "check":
		os.Exit(runCheck())
	case "allow":
		os.Exit(runAllow(os.Args[2:]))
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n", os.Args[1])
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Fprintln(os.Stderr, `Usage: chcc <command>

Commands:
  check          Check chained commands from stdin (Claude Code hook)
  allow          Manage allow list

Run "chcc allow --help" for allow list subcommands.`)
}
