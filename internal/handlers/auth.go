package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lispa/todo-app/internal/middleware"
	"github.com/lispa/todo-app/internal/models"
	"golang.org/x/crypto/bcrypt"
)

// HandleSignup processes new user registration
func HandleSignup(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var u models.User
		if err := json.NewDecoder(r.Body).Decode(&u); err != nil {
			http.Error(w, "Invalid request body", http.StatusBadRequest)
			return
		}

		// Hash password before saving to database
		hashedPassword, _ := bcrypt.GenerateFromPassword([]byte(u.PasswordHash), bcrypt.DefaultCost)

		query := `INSERT INTO users (first_name, last_name, email, password_hash) 
                  VALUES ($1, $2, $3, $4) RETURNING id`
		err := db.QueryRow(r.Context(), query,
			u.FirstName, u.LastName, u.Email, string(hashedPassword)).Scan(&u.ID)

		if err != nil {
			http.Error(w, "User already exists or DB error", http.StatusConflict)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]int{"user_id": u.ID})
	}
}

// HandleLogin authenticates user and returns a JWT token
func HandleLogin(db *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var creds struct {
			Email    string `json:"email"`
			Password string `json:"password"`
		}

		if err := json.NewDecoder(r.Body).Decode(&creds); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		var u models.User
		query := `SELECT id, password_hash FROM users WHERE email = $1`
		err := db.QueryRow(r.Context(), query, creds.Email).Scan(&u.ID, &u.PasswordHash)

		// Compare hashed password with provided plain text
		if err != nil || bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(creds.Password)) != nil {
			http.Error(w, "Invalid email or password", http.StatusUnauthorized)
			return
		}

		// Generate JWT token
		expirationTime := time.Now().Add(24 * time.Hour)
		claims := &models.Claims{
			UserID: u.ID,
			RegisteredClaims: jwt.RegisteredClaims{
				ExpiresAt: jwt.NewNumericDate(expirationTime),
			},
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
		tokenString, _ := token.SignedString(middleware.GetJWTKey())

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
	}
}
