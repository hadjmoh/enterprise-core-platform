package main

import (
	"fmt"
	"os"

	"github.com/hadjmoh/enterprise-core-platform/sdk/pkg/scaffold"
	"github.com/hadjmoh/enterprise-core-platform/sdk/pkg/validator"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	switch command {
	case "init":
		handleInit(os.Args[2:])
	case "validate":
		handleValidate(os.Args[2:])
	case "package":
		handlePackage(os.Args[2:])
	default:
		fmt.Printf("Unknown command: %s\n", command)
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Prospect CLI - Enterprise Core Platform App SDK")
	fmt.Println("\nUsage:")
	fmt.Println("  prospect init <app-name>    Initialize a new app")
	fmt.Println("  prospect validate <path>    Validate an app manifest")
	fmt.Println("  prospect package <path>     Bundle app for deployment")
}

func handleInit(args []string) {
	if len(args) < 1 {
		fmt.Println("Usage: prospect init <app-name>")
		return
	}
	appName := args[0]
	if err := scaffold.CreateApp(appName); err != nil {
		fmt.Printf("Error initializing app: %v\n", err)
		os.Exit(1)
	}
	fmt.Printf("Successfully initialized app '%s'\n", appName)
}

func handleValidate(args []string) {
	path := "."
	if len(args) > 0 {
		path = args[0]
	}
	if err := validator.ValidateApp(path); err != nil {
		fmt.Printf("Validation failed: %v\n", err)
		os.Exit(1)
	}
	fmt.Println("App validation passed successfully!")
}

func handlePackage(args []string) {
	fmt.Println("Packaging not implemented yet.")
}
