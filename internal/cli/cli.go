package cli

import (
	"fmt"
)

func Execute(baseURL string, args []string) error {
	if len(args) < 1 {
		printUsage()
		return nil
	}

	command := args[0]
	commandArgs := args[1:]

	client := NewClient(baseURL)

	switch command {
	case "blacklist":
		return HandleBlacklistCommand(client, commandArgs)
	case "whitelist":
		return HandleWhitelistCommand(client, commandArgs)
	case "reset":
		return HandleResetCommand(client, commandArgs)
	case "help", "--help", "-h":
		printUsage()
		return nil
	default:
		return fmt.Errorf("unknown command: %s. Use 'help' for usage", command)
	}
}

func printUsage() {
	fmt.Println(`Anti-Brute Force CLI

Usage:
  cli [flags] <command> [arguments]

Flags:
  -url string    Base URL of the anti-brute force service (default "http://localhost:8080")

Commands:
  blacklist
    add <cidr>     Add subnet to blacklist
    remove <cidr>  Remove subnet from blacklist
    list           List all subnets in blacklist

  whitelist
    add <cidr>     Add subnet to whitelist
    remove <cidr>  Remove subnet from whitelist
    list           List all subnets in whitelist

  reset [--login <login>] [--ip <ip>]
                   Reset rate limit buckets

  help             Show this help message

Examples:
  cli -url http://localhost:8080 blacklist add 192.168.1.0/24
  cli blacklist list
  cli whitelist add 10.0.0.0/8
  cli reset --login user1
  cli reset --ip 192.168.1.100
  cli reset --login user1 --ip 192.168.1.100`)
}
