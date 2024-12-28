package main

import (
	"ExeProcessManager/command"
	"fmt"
)

func main() {
	// Print a welcome message with more details about the application
	printWelcomeMessage()

	// Start the Command Line Interface (CLI)
	command.StartCLI()

	// Gracefully exit the application
	fmt.Println("Exiting ExeProcessManager. Goodbye!")
}

// printWelcomeMessage outputs a detailed welcome message with application info
func printWelcomeMessage() {
	fmt.Println("=======================================")
	fmt.Println("Welcome to ExeProcessManager!")
	fmt.Println("A process management tool for monitoring and controlling system processes.")
	fmt.Println("Use 'help' for a list of available commands.")
	fmt.Println("=======================================")
}
