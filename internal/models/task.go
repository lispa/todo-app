package models

import "time"

// Task represents the structure of a todo item on your Kanban board.
// We use pointers (*time.Time) for fields that can be NULL in the database.
type Task struct {
	ID         int        `json:"id"`
	UserID     int        `json:"user_id"`
	Title      string     `json:"title"`
	Status     string     `json:"status"` // expected: todo, in_progress, done
	StartedAt  *time.Time `json:"started_at,omitempty"`
	FinishedAt *time.Time `json:"finished_at,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
}
