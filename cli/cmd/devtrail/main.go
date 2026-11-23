package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && (os.Args[1] == "-h" || os.Args[1] == "--help" || os.Args[1] == "help") {
		printUsage()
		return
	}

	printUsage()
	os.Exit(0)
}

func printUsage() {
	fmt.Println("DevTrail CLI")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  devtrail <command> [args]")
	fmt.Println()
	fmt.Println("Commands:")
	fmt.Println("  help     Show this help message")
}
