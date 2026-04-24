package main

import (
	"fmt"
	"os"

	"github.com/fezcode/atlas.clock/pkg/ui"
)

var Version = "dev"

func printHelp() {
	fmt.Println("Atlas Clock — phosphor-CRT TUI for multi-timezone dashboards.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("  atlas.clock          Start the clock dashboard")
	fmt.Println("  atlas.clock -v       Show version")
	fmt.Println("  atlas.clock -h       Show this help")
	fmt.Println()
	fmt.Println("Inside the UI:")
	fmt.Println("  ↑↓←→/hjkl    navigate the grid")
	fmt.Println("  SHIFT+arrow  reorder the selected clock")
	fmt.Println("  ↵            open the detail view")
	fmt.Println("  a            add a clock (label → zone → confirm)")
	fmt.Println("  d            delete the selected clock")
	fmt.Println("  q            quit")
	fmt.Println()
	fmt.Println("Config: ~/.atlas/clock.json")
}

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "-v", "--version":
			fmt.Printf("atlas.clock v%s\n", Version)
			return
		case "-h", "--help", "help":
			printHelp()
			return
		}
	}

	if err := ui.Start(ui.Config{Version: Version}); err != nil {
		fmt.Printf("Error starting UI: %v\n", err)
		os.Exit(1)
	}
}
