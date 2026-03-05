package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5" // Now this will be used!
	"github.com/lispa/todo-app/internal/database"
	"golang.org/x/crypto/bcrypt"
)

// Secret key for JWT signing - keep this safe!
var jwtKey = []byte("my_secret_key_2026")

// User model matching our new database schema
type User struct {
	ID           int       `json:"id"`
	FirstName    string    `json:"first_name"`
	LastName     string    `json:"last_name"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"password"` // Used for input only
	CreatedAt    time.Time `json:"created_at"`
}

// Claims for our JWT token
type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func main() {
	fmt.Println("🚀 Starting Todo-App API with Login support...")

	conn, err := database.Connect()
	if err != nil {
		fmt.Printf("❌ DB Error: %v\n", err)
		return
	}
	defer conn.Close(context.Background())

	// --- HANDLERS ---

	// Registration (SignUp)
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
			http.Error(w, "User already exists", http.StatusConflict)
			return
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{"user_id": u.ID})
	})

	// Login: Verifies credentials and returns a JWT token
	http.HandleFunc("/auth/login", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}
		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		var u User
		// Find user by email
		query := `SELECT id, password_hash FROM users WHERE email = $1`
		err := conn.QueryRow(context.Background(), query, creds.Email).Scan(&u.ID, &u.PasswordHash)

		// Compare password hash
		if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(creds.Password)) != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Create JWT token
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &Claims{
			UserID: u.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, err := token.SignedString(jwtKey)

		if err != nil {
			http.Error(w, "Failed to create token", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	})

	fmt.Println("🌐 Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}
