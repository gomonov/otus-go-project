package cli

import (
	"fmt"
)

func HandleBlacklistCommand(client *Client, args []string) error {
	return handleListCommand(
		args,
		"blacklist",
		client.AddToBlacklist,
		client.RemoveFromBlacklist,
		client.GetBlacklist,
	)
}

func HandleWhitelistCommand(client *Client, args []string) error {
	return handleListCommand(
		args,
		"whitelist",
		client.AddToWhitelist,
		client.RemoveFromWhitelist,
		client.GetWhitelist,
	)
}

func handleListCommand(
	args []string,
	listType string,
	addFunc func(string) error,
	removeFunc func(string) error,
	getFunc func() (*SubnetsListResponse, error),
) error {
	if len(args) < 1 {
		return fmt.Errorf("%s command requires subcommand: add, remove, list", listType)
	}

	subcommand := args[0]
	switch subcommand {
	case "add":
		if len(args) < 2 {
			return fmt.Errorf("%s add requires CIDR argument", listType)
		}
		if err := addFunc(args[1]); err != nil {
			return err
		}
		fmt.Printf("Added %s to %s\n", args[1], listType)
		return nil

	case "remove":
		if len(args) < 2 {
			return fmt.Errorf("%s remove requires CIDR argument", listType)
		}
		if err := removeFunc(args[1]); err != nil {
			return err
		}
		fmt.Printf("Removed %s from %s\n", args[1], listType)
		return nil

	case "list":
		response, err := getFunc()
		if err != nil {
			return err
		}
		fmt.Printf("%s (%d subnets):\n", listType, response.Count)
		for _, subnet := range response.Subnets {
			fmt.Printf("  - %s\n", subnet.CIDR)
		}
		return nil

	default:
		return fmt.Errorf("unknown %s subcommand: %s", listType, subcommand)
	}
}

func HandleResetCommand(client *Client, args []string) error {
	var login, ip string

	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--login", "-l":
			if i+1 < len(args) {
				login = args[i+1]
				i++
			} else {
				return fmt.Errorf("--login requires a value")
			}
		case "--ip", "-i":
			if i+1 < len(args) {
				ip = args[i+1]
				i++
			} else {
				return fmt.Errorf("--ip requires a value")
			}
		default:
			return fmt.Errorf("unknown flag: %s", args[i])
		}
	}

	if login == "" && ip == "" {
		return fmt.Errorf("reset command requires either --login or --ip")
	}

	response, err := client.ResetBuckets(login, ip)
	if err != nil {
		return err
	}

	if response.Reset {
		fmt.Printf("Buckets reset successfully")
		if login != "" {
			fmt.Printf(" for login '%s'", login)
		}
		if ip != "" {
			fmt.Printf(" for IP '%s'", ip)
		}
		fmt.Println()
	} else {
		fmt.Printf("Failed to reset buckets\n")
	}
	return nil
}
