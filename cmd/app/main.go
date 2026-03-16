package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lispa/todo-app/internal/database"
	"github.com/lispa/todo-app/internal/handlers"
	"github.com/lispa/todo-app/internal/middleware"
)

func main() {
	fmt.Println("🚀 Starting Refactored Todo-App API Server...")

	// 1. Establish database connection with retry logic
	conn := connectWithRetry()
	defer conn.Close(context.Background())

	// 2. Define Routes

	// Auth Routes
	http.HandleFunc("/auth/signup", middleware.EnableCORS(handlers.HandleSignup(conn)))
	http.HandleFunc("/auth/login", middleware.EnableCORS(handlers.HandleLogin(conn)))

	// Task Routes (Protected by Auth Middleware)
	http.HandleFunc("/tasks", middleware.EnableCORS(middleware.Auth(handlers.HandleListTasks(conn))))
	http.HandleFunc("/tasks/create", middleware.EnableCORS(middleware.Auth(handlers.HandleCreateTask(conn))))
	http.HandleFunc("/tasks/start", middleware.EnableCORS(middleware.Auth(handlers.HandleStartTask(conn))))
	http.HandleFunc("/tasks/done", middleware.EnableCORS(middleware.Auth(handlers.HandleDoneTask(conn))))
	http.HandleFunc("/tasks/delete", middleware.EnableCORS(middleware.Auth(handlers.HandleDeleteTask(conn))))

	// 3. Start Server
	fmt.Println("🌐 Server listening on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("❌ Server failed to start: %v\n", err)
	}
}

// connectWithRetry attempts to connect to the database 10 times before failing
func connectWithRetry() *pgx.Conn {
	var conn *pgx.Conn
	var err error

	for i := 1; i <= 10; i++ {
		conn, err = database.Connect()
		if err == nil {
			fmt.Println("✅ Successfully connected to the database!")
			return conn
		}
		fmt.Printf("⏳ [Attempt %d/10] Database not ready, retrying in 3 seconds...\n", i)
		time.Sleep(3 * time.Second)
	}

	panic(fmt.Sprintf("❌ Critical Error: Could not connect to DB after 10 attempts: %v", err))
}
