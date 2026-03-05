package main

import (
	"context"
	"fmt"
	"time"

	"internal/database"
)

func main() {
	fmt.Println("🚀 Starting Todo-App...")

	// Trying to connect
	var conn interface{} // Let's simplify it for the sake of example, in reality there is *pgx.Conn

	// We'll use the waitForDB logic, but expand it a bit.
	for {
		c, err := database.Connect()
		if err != nil {
			fmt.Printf("⚠️ Waiting for DB: %v\n", err)
			time.Sleep(2 * time.Second)
			continue
		}
		fmt.Println("✅ Connected and Table is ready!")

		// Let's insert a test task
		var id int
		err = c.QueryRow(context.Background(),
			"INSERT INTO tasks (title) VALUES ($1) RETURNING id",
			"My first task from Go!").Scan(&id)

		if err != nil {
			fmt.Printf("❌ Failed to insert task: %v\n", err)
		} else {
			fmt.Printf("📝 Inserted task with ID: %d\n", id)
		}

		c.Close(context.Background())
		break
	}

	fmt.Println("🚀 Application finished its work.")
}
