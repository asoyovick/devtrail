package cmd

import "fmt"

func GitHub(args []string) error {
	if len(args) < 1 { return fmt.Errorf("usage: devtrail github <connect|status>") }
	switch args[0] {
	case "connect":
		fmt.Println("GitHub connect is not yet implemented in this build. Use the web integrations page once backend GitHub OAuth is configured.")
		return nil
	case "status":
		fmt.Println("GitHub status: not connected (GitHub integration pending).")
		return nil
	default:
		return fmt.Errorf("unknown github command: %s", args[0])
	}
}
