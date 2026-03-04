package main

import (
	"fmt"
	"time"
)

func main() {
	fmt.Println("🚀 Todo App successfully launched in Docker on Proxmox!")

	// A temporary plug to prevent the container from closing
	for {
		fmt.Printf("The server is running... Current time: %s\n", time.Now().Format("15:04:05"))
		time.Sleep(10 * time.Minute)
	}
}
