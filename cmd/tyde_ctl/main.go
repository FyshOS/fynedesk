package main

import (
	"fmt"
	"net/rpc"
	"os"
	"strings"

	"fyshos.com/tyde"
	frpc "fyshos.com/tyde/modules/rpc"
)

func main() {
	if len(os.Args) < 2 || os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help" {
		printHelp()
		return
	}

	client, err := rpc.Dial("unix", frpc.SocketPath())
	if err != nil {
		fmt.Fprintln(os.Stderr, "failed to connect to tyde:", err)
		os.Exit(1)
	}
	defer client.Close()

	var commands []tyde.Command
	if err := client.Call("Service.ListCommands", "", &commands); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if os.Args[1] == "list" {
		fmt.Println("Available commands:")
		for _, c := range commands {
			fmt.Println("  " + usage(c))
			fmt.Println("      " + c.Description)
		}
		return
	}

	for _, cmd := range commands {
		if cmd.Name != os.Args[1] {
			continue
		}

		var reply string
		if err := client.Call("Service.Launch", strings.Join(os.Args[1:], " "), &reply); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Println(reply)
		return
	}

	fmt.Fprintln(os.Stderr, "Could not find command matching:", os.Args[1])
	os.Exit(1)
}

// usage writes a command's invocation line, e.g. "brightness <a percentage …>".
func usage(c tyde.Command) string {
	if c.Args == "" {
		return c.Name
	}
	return c.Name + " [" + c.Args + "]"
}

func printHelp() {
	fmt.Print(`tyde_ctl - command line interface for Tyde

Usage:
  tyde_ctl <command> [args...]
  tyde_ctl list
  tyde_ctl help

Commands available depend on loaded modules - "tyde_ctl list" shows those available.

Examples:
  tyde_ctl brightness up       Increase screen brightness
  tyde_ctl brightness 50       Set brightness to 50%
  tyde_ctl desktop 2           Switch to the second virtual desktop
  tyde_ctl send Firefox to 3   Move the Firefox window to the third desktop
  tyde_ctl windows             List the open windows
  tyde_ctl focus Firefox       Bring the Firefox window to the front
  tyde_ctl maximize            Maximize the focused window
  tyde_ctl move Firefox left   Tile the Firefox window to the left half
  tyde_ctl resize 1280 720     Resize the focused window
  tyde_ctl big Hello World     Show "Hello World" in large type

Built-in commands:
  list     Show the commands this desktop offers, with their arguments
  help     Show this help message
`)
}
