package main

import (
	"fmt"
	"log"
	"systemd-tui/internal/service"
)

func main() {
	client := service.NewSystemdClient()
	units, err := client.ListUnits()
	if err != nil {
		log.Fatalf("Error listing units: %v", err)
	}

	fmt.Printf("Found %d user services:\n", len(units))
	for _, u := range units {
		fmt.Printf("- %-30s [%s] %s\n", u.Unit, u.Active, u.Description)
	}
}
