package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/lispa/todo-app/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// User represents the account owner in the system
type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"` // We use this field for input, then hash it
	CreatedAt    time.Time `json:"created_at"`
}

// Task represents a todo item with lifecycle tracking
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
	fmt.Println("🚀 Starting Todo-App API Server...")

	// Initialize database connection
	conn, err := database.Connect()
	if err != nil {
		fmt.Printf("❌ Database error: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	fmt.Println("✅ Database connected")

	// --- ROUTES ---

	// User Registration Handler
	http.HandleFunc("/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		// Only allow POST requests for security
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Hash password before storing it
		hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		if err != nil {
			http.Error(w, "Server error during password hashing", http.StatusInternalServerError)
			return
		}

		// Insert user and get the new ID
		query := `
			INSERT INTO users (first_name, last_name, email, password_hash)
			VALUES ($1, $2, $3, $4)
			RETURNING id, created_at`

		err = conn.QueryRow(context.Background(), query,
			u.FirstName, u.LastName, u.Email, string(hashedPassword)).Scan(&u.ID, &u.CreatedAt)

		if err != nil {
			fmt.Printf("DB Error: %v\n", err)
			http.Error(w, "User already exists or database error", http.StatusConflict)
			return
		}

		// Success response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "User registered successfully",
			"user_id": u.ID,
		})
	})

	// Health check
	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("pong"))
	})

	fmt.Println("🌐 Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("❌ Server crash: %v\n", err)
	}
}
