package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lispa/todo-app/internal/database"
)

// Task defines the structure for our JSON data
type Task struct {
	ID        int       `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}

func main() {
	fmt.Println("🚀 Starting Todo-App API server...")

	// Initialize database connection
	conn, err := database.Connect()
	if err != nil {
		fmt.Printf("❌ Connection error: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	fmt.Println("✅ Database connection verified")

	// Endpoint to get all tasks as JSON
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		rows, err := conn.Query(context.Background(), "SELECT id, title, done, created_at FROM tasks ORDER BY id DESC")
		if err != nil {
			http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var t Task
			if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
				continue
			}
			tasks = append(tasks, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	})

	// Basic health check endpoint
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	fmt.Println("🌐 Server is listening on port :8080")
	// ListenAndServe blocks the app and keeps it running
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
	}
}
