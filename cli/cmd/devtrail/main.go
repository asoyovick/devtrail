package main

import (
	"fmt"
	"os"

	icmd "devtrail/cli/internal/cmd"
)

func main() {
	if len(os.Args) < 2 {
		icmd.PrintUsage()
		return
	}
	cmd := os.Args[1]
	args := os.Args[2:]
	var err error
	switch cmd {
	case "help", "-h", "--help":
		icmd.PrintUsage()
	case "init":
		err = icmd.Init(args)
	case "login":
		err = icmd.Login(args)
	case "connect":
		err = icmd.Connect(args)
	case "disconnect":
		err = icmd.Disconnect(args)
	case "projects":
		err = icmd.Projects(args)
	case "status":
		err = icmd.Status(args)
	case "watch":
		err = icmd.Watch(args)
	case "sync":
		err = icmd.Sync(args)
	case "session":
		err = icmd.Session(args)
	case "sessions":
		err = icmd.Sessions(args)
	case "github":
		err = icmd.GitHub(args)
	default:
		fmt.Fprintf(os.Stderr, "unknown command: %s\n\n", cmd)
		icmd.PrintUsage()
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}
