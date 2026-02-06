package main

import (
	"fmt"
	"log"
)

func main() {
	// Example usage of the config loader
	// In a real application, this would be used to initialize the TUI
	fmt.Println("o8n - Terminal UI for Operaton")
	fmt.Println("-------------------------------")
	
	// Try to load config if it exists
	config, err := LoadConfig("config.yaml")
	if err != nil {
		log.Printf("Note: Could not load config.yaml: %v", err)
		log.Printf("You can create a config.yaml based on config.yaml.example")
		return
	}
	
	// Display loaded environments
	fmt.Printf("\nLoaded %d environment(s):\n", len(config.Environments))
	for name, env := range config.Environments {
		fmt.Printf("  - %s: %s (color: %s)\n", name, env.URL, env.UIColor)
	}
}
