package main

import (
	"context"
	"fmt"
	"time"

	"github.com/lispa/todo-app/internal/database"
)

func main() {
	fmt.Println("🚀 Starting the process of connecting to the database...")

	for {
		// 1. Connect and create a table
		conn, err := database.Connect()
		if err != nil {
			fmt.Printf("⚠️ Database not yet available: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("✅ Connection established and table is ready!")

		// 2. USE conn to insert a test task
		var id int
		err = conn.QueryRow(context.Background(),
			"INSERT INTO tasks (title) VALUES ($1) RETURNING id",
			"My first real task!").Scan(&id)

		if err != nil {
			fmt.Printf("❌ Failed to insert task: %v\n", err)
		} else {
			fmt.Printf("📝 Success! Created task with ID: %d\n", id)
		}

		// 3. Close the connection before exiting
		defer conn.Close(context.Background())
		break
	}

	fmt.Println("🚀 The application has finished its job.")
}
