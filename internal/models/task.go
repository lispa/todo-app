package repository

import (
	"context"

	"github.com/lispa/todo-app/internal/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TaskRepository struct {
	Pool *pgxpool.Pool
}

// NewTaskRepository creates a new instance of TaskRepository
func NewTaskRepository(pool *pgxpool.Pool) *TaskRepository {
	return &TaskRepository{Pool: pool}
}

// UpdateStatus changes the status and sets timestamps
func (r *TaskRepository) UpdateStatus(ctx context.Context, taskID int, newStatus string) error {
	var query string
	switch newStatus {
	case "in_progress":
		query = "UPDATE tasks SET status = $1, started_at = NOW() WHERE id = $2"
	case "done":
		query = "UPDATE tasks SET status = $1, finished_at = NOW() WHERE id = $2"
	default:
		query = "UPDATE tasks SET status = $1 WHERE id = $2"
	}
	_, err := r.Pool.Exec(ctx, query, newStatus, taskID)
	return err
}

// GetAll returns all tasks for the board
func (r *TaskRepository) GetAll(ctx context.Context) ([]models.Task, error) {
	query := "SELECT id, title, status, started_at, finished_at FROM tasks"
	rows, err := r.Pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []models.Task
	for rows.Next() {
		var t models.Task
		// Scan data into the struct
		err := rows.Scan(&t.ID, &t.Title, &t.Status, &t.StartedAt, &t.FinishedAt)
		if err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}
