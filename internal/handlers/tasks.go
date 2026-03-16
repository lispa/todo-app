package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/lispa/todo-app/internal/models"
)

// HandleListTasks retrieves all tasks for the authenticated user
func HandleListTasks(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value("user_id").(int)

		query := `SELECT id, user_id, title, status, started_at, finished_at, created_at 
		          FROM tasks WHERE user_id = $1 ORDER BY created_at DESC`

		rows, err := db.Query(context.Background(), query, userID)
		if err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		tasks := []models.Task{}
		for rows.Next() {
			var t models.Task
			err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt)
			if err != nil {
				continue
			}
			tasks = append(tasks, t)
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

// HandleCreateTask adds a new task to the database
func HandleCreateTask(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		userID := r.Context().Value("user_id").(int)
		var t models.Task
		if err := json.NewDecoder(r.Body).Decode(&t); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		query := `INSERT INTO tasks (user_id, title, status) 
		          VALUES ($1, $2, $3) RETURNING id, user_id, status, created_at`

		err := db.QueryRow(context.Background(), query, userID, t.Title, "todo").
			Scan(&t.ID, &t.UserID, &t.Status, &t.CreatedAt)

		if err != nil {
			http.Error(w, "Failed to create task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(t)
	}
}

// HandleStartTask updates task status to 'in_progress'
func HandleStartTask(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)

		var startedAt time.Time
		query := `UPDATE tasks SET status = 'in_progress', started_at = CURRENT_TIMESTAMP 
		          WHERE id = $1 AND user_id = $2 RETURNING started_at`

		err := db.QueryRow(context.Background(), query, input.ID, userID).Scan(&startedAt)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":     "in_progress",
			"started_at": startedAt,
		})
	}
}

// HandleDoneTask marks task as completed
func HandleDoneTask(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)

		var finishedAt time.Time
		query := `UPDATE tasks SET status = 'done', finished_at = CURRENT_TIMESTAMP 
		          WHERE id = $1 AND user_id = $2 RETURNING finished_at`

		err := db.QueryRow(context.Background(), query, input.ID, userID).Scan(&finishedAt)
		if err != nil {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"status":      "done",
			"finished_at": finishedAt,
		})
	}
}

// HandleDeleteTask removes a task from the database
func HandleDeleteTask(db *pgx.Conn) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodDelete {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var input struct {
			ID int `json:"id"`
		}
		json.NewDecoder(r.Body).Decode(&input)
		userID := r.Context().Value("user_id").(int)

		res, err := db.Exec(context.Background(), "DELETE FROM tasks WHERE id = $1 AND user_id = $2", input.ID, userID)
		if err != nil || res.RowsAffected() == 0 {
			http.Error(w, "Task not found", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
