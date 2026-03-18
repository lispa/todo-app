package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lispa/todo-app/internal/models"
)

// TaskRepository handles all database operations for tasks
type TaskRepository struct {
	Pool *pgxpool.Pool
}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{Pool: pool}
}

// GetAll returns all tasks for a specific user
// This will feed your "TODO", "In Progress", and "Done" columns
func (r *TaskRepository) GetAll(ctx context.Context, userID int) ([]models.Task, error) {
	query := `
		SELECT id, user_id, title, status, started_at, finished_at, created_at 
		FROM tasks 
		WHERE user_id = $1 
		ORDER BY created_at DESC`

	rows, err := r.Pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("query failed: %v", err)
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		// We use pointers in models for nullable fields like started_at
		err := rows.Scan(&t.ID, &t.UserID, &t.Title, &t.Status, &t.StartedAt, &t.FinishedAt, &t.CreatedAt)
		if err != nil {
			return nil, fmt.Errorf("scan failed: %v", err)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

// Create adds a new task with default 'todo' status
func (r *TaskRepository) Create(ctx context.Context, userID int, title string) (models.Task, error) {
	var t models.Task
	query := `
		INSERT INTO tasks (user_id, title, status) 
		VALUES ($1, $2, 'todo') 
		RETURNING id, user_id, status, created_at`

	err := r.Pool.QueryRow(ctx, query, userID, title).
		Scan(&t.ID, &t.UserID, &t.Status, &t.CreatedAt)

	t.Title = title // Set the title since RETURNING didn't include it
	return t, err
}

// UpdateStatus moves a task to a different column and manages timestamps
func (r *TaskRepository) UpdateStatus(ctx context.Context, userID, taskID int, newStatus string) error {
	var query string

	// Kanban Logic: handle timestamps based on status transitions
	switch newStatus {
	case "in_progress":
		query = "UPDATE tasks SET status = $1, started_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3"
	case "done":
		query = "UPDATE tasks SET status = $1, finished_at = CURRENT_TIMESTAMP WHERE id = $2 AND user_id = $3"
	default:
		// Moving back to 'todo'
		query = "UPDATE tasks SET status = $1 WHERE id = $2 AND user_id = $3"
	}

	res, err := r.Pool.Exec(ctx, query, newStatus, taskID, userID)
	if err != nil {
		return err
	}

	if res.RowsAffected() == 0 {
		return fmt.Errorf("task not found or access denied")
	}
	return nil
}

// Delete removes a task from the board
func (r *TaskRepository) Delete(ctx context.Context, userID, taskID int) error {
	res, err := r.Pool.Exec(ctx, "DELETE FROM tasks WHERE id = $1 AND user_id = $2", taskID, userID)
	if err != nil {
		return err
	}
	if res.RowsAffected() == 0 {
		return fmt.Errorf("task not found")
	}
	return nil
}
