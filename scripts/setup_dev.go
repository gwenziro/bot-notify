package main

import (
	"fmt"
	"os"
	"os/exec"
	"runtime"
)

func main() {
	fmt.Println("Setting up development environment for WhatsApp Bot...")

	// Install Air for hot-reloading
	fmt.Println("Installing Air - hot-reloading tool for Go...")
	cmd := exec.Command("go", "install", "github.com/cosmtrek/air@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		fmt.Printf("Error installing Air: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("\nAir installed successfully!")
	fmt.Println("\nTo start development with hot-reload, run:")

	if runtime.GOOS == "windows" {
		fmt.Println("air.exe")
	} else {
		fmt.Println("air")
	}

	fmt.Println("\nSetup complete! Happy coding!")
}
