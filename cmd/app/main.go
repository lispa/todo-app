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

// Secret key for JWT signing - in production, use environment variables!
var jwtKey = []byte("my_ultra_secret_key_2026")

// User represents the account owner details
type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"` // Used for input, hidden in output
	CreatedAt    time.Time `json:"created_at"`
}

// Task represents a todo item linked to a user
type Task struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"` // todo, in_progress, done
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}

// Claims defines the structure of the JWT payload
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

// authMiddleware verifies the JWT token before allowing access to a route
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

		// Pass the user_id to the next handler via context
		ctx := context.WithValue(r.Context(), "user_id", claims.UserID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func main() {
	fmt.Println("🚀 Starting Todo-App API Server...")

	conn, err := database.Connect()
	if err != nil {
		fmt.Printf("❌ Database connection failed: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	fmt.Println("✅ Connected to PostgreSQL")

	// --- PUBLIC ROUTES ---

	// SIGNUP: Register a new John Doe
	http.HandleFunc("/auth/signup", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST method required", http.StatusMethodNotAllowed)
			return
		}
		var u User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)
		query := `INSERT INTO users (first_name, last_name, email, password_hash) 
                  VALUES ($1, $2, $3, $4) RETURNING id, created_at`

		err := conn.QueryRow(context.Background(), query,
			u.FirstName, u.LastName, u.Email, string(hashedPassword)).Scan(&u.ID, &u.CreatedAt)

		if err != nil {
			http.Error(w, "Registration failed (email might exist)", http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(u)
	})

	// LOGIN: Verify user and return a JWT
	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST method required", http.StatusMethodNotAllowed)
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
		claims := &Claims{
			UserID: u.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(jwtKey)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	})

	// --- PROTECTED ROUTES (Require Token) ---

	// CREATE TASK: Add a task for the authenticated user
	http.HandleFunc("/tasks/create", authMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "POST method required", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value("user_id").(int)
		var t Task
		json.NewDecoder(r.Body).Decode(&t)

		query := `INSERT INTO tasks (user_id, title, status) VALUES ($1, $2, $3) RETURNING id, created_at`
		err := conn.QueryRow(context.Background(), query, userID, t.Title, "todo").Scan(&t.ID, &t.CreatedAt)

		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(t)
	}))

	fmt.Println("🌐 Server is running on port :8080")
	http.ListenAndServe(":8080", nil)
}
