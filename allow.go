package main

import (
	"fmt"
	"os"
	"strings"
)

func runAllow(args []string) int {
	if len(args) == 0 {
		printAllowUsage()
		return 1
	}

	switch args[0] {
	case "list":
		return allowList()
	case "add":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: chcc allow add <command>")
			return 1
		}
		return allowAdd(strings.Join(args[1:], " "))
	case "remove":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: chcc allow remove <command>")
			return 1
		}
		return allowRemove(strings.Join(args[1:], " "))
	default:
		fmt.Fprintf(os.Stderr, "unknown allow subcommand: %s\n", args[0])
		printAllowUsage()
		return 1
	}
}

func printAllowUsage() {
	fmt.Fprintln(os.Stderr, `Usage: chcc allow <subcommand>

Subcommands:
  list                 List allowed commands
  add <command>        Add a command to the allow list
  remove <command>     Remove a command from the allow list`)
}

func allowList() int {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	if len(cfg.AllowList) == 0 {
		fmt.Println("(empty)")
		return 0
	}
	for _, cmd := range cfg.AllowList {
		fmt.Println(cmd)
	}
	return 0
}

func allowAdd(cmd string) int {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	for _, existing := range cfg.AllowList {
		if existing == cmd {
			fmt.Fprintf(os.Stderr, "already in allow list: %s\n", cmd)
			return 0
		}
	}

	cfg.AllowList = append(cfg.AllowList, cmd)
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("added: %s\n", cmd)
	return 0
}

func allowRemove(cmd string) int {
	cfg, err := loadConfig()
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}

	found := false
	filtered := make([]string, 0, len(cfg.AllowList))
	for _, existing := range cfg.AllowList {
		if existing == cmd {
			found = true
			continue
		}
		filtered = append(filtered, existing)
	}

	if !found {
		fmt.Fprintf(os.Stderr, "not in allow list: %s\n", cmd)
		return 1
	}

	cfg.AllowList = filtered
	if err := saveConfig(cfg); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		return 1
	}
	fmt.Printf("removed: %s\n", cmd)
	return 0
}
