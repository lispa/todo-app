package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lispa/todo-app/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// Secret key for JWT signing
var jwtKey = []byte("my_ultra_secret_key_2026")

// User model
type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"`
	CreatedAt    time.Time `json:"created_at"`
}

// Task model with time tracking
type Task struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Claims for JWT
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// authMiddleware to protect routes
func authMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tokenString := r.Header.Get("Authorization")
		if tokenString == "" {
			http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
			return
		}

		claims := &Claims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtKey, nil
		})

		if err != nil || !token.Valid {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func main() {
	fmt.Println("🚀 Starting Todo-App API Server...")

	conn, err := database.Connect()
	if err != nil {
		fmt.Printf("❌ DB Error: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	// --- AUTH ROUTES ---

	http.HandleFunc("/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var u User
		json.NewDecoder(r.Body).Decode(&u)
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)

		query := `INSERT INTO users (first_name, last_name, email, password_hash) VALUES ($1, $2, $3, $4) RETURNING id`
		err := conn.QueryRow(context.Background(), query, u.FirstName, u.LastName, u.Email, string(hashedPassword)).Scan(&u.ID)
		if err != nil {
			http.Error(w, "Registration failed", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]int{"user_id": u.ID})
	})

	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		json.NewDecoder(r.Body).Decode(&creds)

		var u User
		query := `SELECT id, password_hash FROM users WHERE email = $1`
		err := conn.QueryRow(context.Background(), query, creds.Email).Scan(&u.ID, &u.PasswordHash)

		if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(creds.Password)) != nil {
			http.Error(w, "Invalid credentials", http.StatusUnauthorized)
			return
		}

		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{UserID: u.ID, RegisteredClaims: jwt.RegisteredClaims{ExpiresAt: jwt.NewNumericDate(expirationTime)}}
		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(jwtKey)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	})

	// --- TASK ROUTES ---

	// Create Task
	http.HandleFunc("/tasks/create", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		userID := r.Context().Value("user_id").(int)
		var t Task
		json.NewDecoder(r.Body).Decode(&t)

		// Fixed: Scan all returning fields to update the struct properly
		query := `INSERT INTO tasks (user_id, title, status) VALUES ($1, $2, $3) RETURNING id, user_id, status, created_at`
		err := conn.QueryRow(context.Background(), query, userID, t.Title, "todo").
			Scan(&t.ID, &t.UserID, &t.Status, &t.CreatedAt)

		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	}))

	// Start Task (Set in_progress)
	http.HandleFunc("/tasks/start", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)

		query := `UPDATE tasks SET status = 'in_progress', started_at = CURRENT_TIMESTAMP 
				  WHERE id = $1 AND user_id = $2 RETURNING started_at`

		var startedAt time.Time
		err := conn.QueryRow(context.Background(), query, input.ID, userID).Scan(&startedAt)
		if err != nil {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "in_progress", "started_at": startedAt})
	}))

	// Finish Task (Set done)
	http.HandleFunc("/tasks/done", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST required", http.StatusMethodNotAllowed)
			return
		}
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)

		query := `UPDATE tasks SET status = 'done', finished_at = CURRENT_TIMESTAMP 
				  WHERE id = $1 AND user_id = $2 RETURNING finished_at`

		var finishedAt time.Time
		err := conn.QueryRow(context.Background(), query, input.ID, userID).Scan(&finishedAt)
		if err != nil {
			http.Error(w, "Task not found or unauthorized", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "done", "finished_at": finishedAt})
	}))

	fmt.Println("🌐 Listening on :8080")
	http.ListenAndServe(":8080", nil)
}
