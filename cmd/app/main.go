package main

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lispa/todo-app/internal/database"
	"github.com/lispa/todo-app/internal/middleware"
)

func main() {
	fmt.Println("🚀 Starting Refactored Todo-App API...")

	// 1. Database Connection (Retry Logic)
	conn := connectWithRetry()
	defer conn.Close(context.Background())

	// 2. Routing
	// Auth
	http.HandleFunc("/auth/signup", middleware.EnableCORS(HandleSignup(conn)))
	http.HandleFunc("/auth/login", middleware.EnableCORS(HandleLogin(conn)))

	// Tasks (Wrapped in Auth Middleware)
	http.HandleFunc("/tasks", middleware.EnableCORS(middleware.Auth(HandleListTasks(conn))))
	http.HandleFunc("/tasks/create", middleware.EnableCORS(middleware.Auth(HandleCreateTask(conn))))

	fmt.Println("🌐 Server listening on :8080")
	http.ListenAndServe(":8080", nil)
}

func connectWithRetry() *pgx.Conn {
	for i := 1; i <= 10; i++ {
		conn, err := database.Connect()
		if err == nil {
			return conn
		}
		time.Sleep(3 * time.Second)
	}
	panic("Could not connect to DB")
}
