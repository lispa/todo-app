package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv" // Don't forget: go get github.com/joho/godotenv
	"github.com/lispa/todo-app/internal/database"
	"github.com/lispa/todo-app/internal/handlers"
	"github.com/lispa/todo-app/internal/middleware"
	"github.com/lispa/todo-app/internal/repository"
)

func main() {
	fmt.Println("🚀 Starting Refactored Kanban Todo-App API Server...")

	// 1. Load environment variables from .env file
	// This ensures os.Getenv works correctly for DB credentials
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	// 2. Establish database connection pool with retry logic
	// Returns *pgxpool.Pool instead of a single connection
	pool := connectWithRetry()
	defer pool.Close()

	// 3. Initialize Repository
	// We wrap the database pool into a repository layer
	taskRepo := repository.NewTaskRepository(pool)

	// 4. Define Routes
	// Note: We pass taskRepo to handlers instead of the raw db connection

	// Auth Routes (Assuming you'll refactor these too later)
	http.HandleFunc("/auth/signup", middleware.EnableCORS(handlers.HandleSignup(pool)))
	http.HandleFunc("/auth/login", middleware.EnableCORS(handlers.HandleLogin(pool)))

	// Task Routes (Protected by Auth Middleware)
	// These handlers now use our new TaskRepository for Kanban logic
	http.HandleFunc("/tasks", middleware.EnableCORS(middleware.Auth(handlers.HandleListTasks(taskRepo))))
	http.HandleFunc("/tasks/create", middleware.EnableCORS(middleware.Auth(handlers.HandleCreateTask(taskRepo))))

	// New Unified Route for moving tasks between columns (TODO -> In Progress -> Done)
	http.HandleFunc("/tasks/update-status", middleware.EnableCORS(middleware.Auth(handlers.HandleUpdateStatus(taskRepo))))

	http.HandleFunc("/tasks/delete", middleware.EnableCORS(middleware.Auth(handlers.HandleDeleteTask(taskRepo))))

	// 5. Start Server
	fmt.Println("🌐 Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
	}
}

// connectWithRetry attempts to connect to the database 10 times before failing
// Returns *pgxpool.Pool which is safe for high-concurrency Kanban boards
func connectWithRetry() *pgxpool.Pool {
	var pool *pgxpool.Pool
	var err error

	for i := 1; i <= 10; i++ {
		// Calling your updated database.Connect() which should return (*pgxpool.Pool, error)
		pool, err = database.Connect()
		if err == nil {
			fmt.Println("✅ Successfully connected to the database pool!")
			return pool
		}
		fmt.Printf("⏳ [Attempt %d/10] Database not ready, retrying in 3 seconds...\n", i)
		time.Sleep(3 * time.Second)
	}

	panic(fmt.Sprintf("❌ Critical Error: Could not connect to DB after 10 attempts: %v", err))
}
