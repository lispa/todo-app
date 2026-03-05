package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lispa/todo-app/internal/database"
)

// User represents the account owner
type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"` // Hidden in JSON
	CreatedAt    time.Time `json:"created_at"`
}

// Task represents the todo item with lifecycle tracking
type Task struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

func main() {
	fmt.Println("🚀 Starting Todo-App API...")

	conn, err := database.Connect()
	if err != nil {
		fmt.Printf("❌ Database error: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	// Endpoint to get all tasks
	http.HandleFunc("/tasks", func(w http.ResponseWriter, r *http.Request) {
		query := "SELECT id, user_id, title, status, started_at, finished_at, created_at FROM tasks ORDER BY id DESC"
		rows, err := conn.Query(context.Background(), query)
		if err != nil {
			http.Error(w, "Query error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		var tasks []Task
		for rows.Next() {
			var t Task
			// Scanning with pointers to handle potential NULL values in time fields
			err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt)
			if err != nil {
				continue
			}
			tasks = append(tasks, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	})

	fmt.Println("🌐 Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
