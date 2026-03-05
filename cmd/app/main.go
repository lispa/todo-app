package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5" // Ensure this matches your database.Connect return type
	"github.com/lispa/todo-app/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// JWT Secret Key - In a real app, use os.Getenv("JWT_SECRET")
var jwtKey = []byte("my_ultra_secret_key_2026")

// --- MODELS ---

type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"`
	CreatedAt    time.Time `json:"created_at"`
}

type Task struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"`
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// --- MIDDLEWARE ---

// enableCORS: Standard CORS headers for browser compatibility
func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, DELETE, PUT")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	}
}

// authMiddleware: Validates JWT from Authorization header
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

// --- MAIN ---

func main() {
	fmt.Println("🚀 Starting Todo-App API Server...")

	var conn *pgx.Conn
	var err error

	// RETRY LOGIC: Wait for PostgreSQL to be ready in Docker
	for i := 1; i <= 10; i++ {
		conn, err = database.Connect()
		if err == nil {
			fmt.Println("✅ Successfully connected to the database!")
			break
		}
		fmt.Printf("⏳ [Attempt %d/10] Database not ready, retrying in 3 seconds...\n", i)
		time.Sleep(3 * time.Second)
	}

	if err != nil {
		fmt.Printf("❌ Critical Error: Could not connect to DB after 10 attempts: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	// --- AUTH ROUTES ---
	http.HandleFunc("/auth/signup", enableCORS(func(w http.ResponseWriter, r *http.Request) {
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
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"user_id": u.ID})
	}))

	http.HandleFunc("/auth/login", enableCORS(func(w http.ResponseWriter, r *http.Request) {
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
	}))

	// --- TASK ROUTES ---

	// List Tasks
	http.HandleFunc("/tasks", enableCORS(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := r.Context().Value("user_id").(int)
		rows, err := conn.Query(context.Background(), "SELECT id, user_id, title, status, started_at, finished_at, created_at FROM tasks WHERE user_id = $1 ORDER BY created_at DESC", userID)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()
		tasks := []Task{}
		for rows.Next() {
			var t Task
			rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt)
			tasks = append(tasks, t)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	})))

	// Create Task
	http.HandleFunc("/tasks/create", enableCORS(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		userID := r.Context().Value("user_id").(int)
		var t Task
		json.NewDecoder(r.Body).Decode(&t)
		query := `INSERT INTO tasks (user_id, title, status) VALUES ($1, $2, $3) RETURNING id, user_id, status, created_at`
		err := conn.QueryRow(context.Background(), query, userID, t.Title, "todo").Scan(&t.ID, &t.UserID, &t.Status, &t.CreatedAt)
		if err != nil {
			http.Error(w, "DB error", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	})))

	// Start Task
	http.HandleFunc("/tasks/start", enableCORS(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)
		var startedAt time.Time
		err := conn.QueryRow(context.Background(), "UPDATE tasks SET status = 'in_progress', started_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 RETURNING started_at", input.ID, userID).Scan(&startedAt)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "in_progress", "started_at": startedAt})
	})))

	// Done Task
	http.HandleFunc("/tasks/done", enableCORS(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)
		var finishedAt time.Time
		err := conn.QueryRow(context.Background(), "UPDATE tasks SET status = 'done', finished_at = CURRENT_TIMESTAMP WHERE id = $1 AND user_id = $2 RETURNING finished_at", input.ID, userID).Scan(&finishedAt)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{"status": "done", "finished_at": finishedAt})
	})))

	// Delete Task
	http.HandleFunc("/tasks/delete", enableCORS(authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)
		res, err := conn.Exec(context.Background(), "DELETE FROM tasks WHERE id = $1 AND user_id = $2", input.ID, userID)
		if err != nil || res.RowsAffected() == 0 {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	})))

	fmt.Println("🌐 Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
