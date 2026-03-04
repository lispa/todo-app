package main

import (
	"fmt"
	"time"

	"github.com/lispa/todo-app/internal/database"
)

func waitForDB() {
	fmt.Println("🚀 Starting the process of connecting to the database...")

	attempt := 1

	for {
		conn, err := database.Connect()
		if err != nil {
			fmt.Printf("⚠️ Attempt %d: Database not yet available...\n", attempt)
			attempt++
			time.Sleep(2 * time.Second)
			continue
		}

		fmt.Println("✅ Hooray! The connection to PostgreSQL has been successfully established!")

		defer conn.Close(conn.Context())

		break
	}
}

func main() {

	waitForDB()
	fmt.Println("🚀 The application has been launched")
}
