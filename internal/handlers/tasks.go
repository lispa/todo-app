package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/lispa/todo-app/internal/repository"
)

// HandleListTasks - GET /tasks
// Fetches all tasks to populate your Kanban columns
func HandleListTasks(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get userID from context (set by your Auth middleware)
		userID, ok := r.Context().Value("user_id").(int)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		tasks, err := repo.GetAll(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to fetch tasks", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(tasks)
	}
}

// HandleCreateTask - POST /tasks/create
func HandleCreateTask(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			Title string `json:"title"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(int)
		task, err := repo.Create(r.Context(), userID, input.Title)
		if err != nil {
			http.Error(w, "Could not create task", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(task)
	}
}

// HandleUpdateStatus - PATCH /tasks/update-status
// This is what happens when you drag a task in your Paint drawing
func HandleUpdateStatus(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID     int    `json:"id"`
			Status string `json:"status"` // todo, in_progress, done
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid input", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(int)
		err := repo.UpdateStatus(r.Context(), userID, input.ID, input.Status)
		if err != nil {
			http.Error(w, "Update failed: "+err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"message": "Status updated successfully"})
	}
}

// HandleDeleteTask - DELETE /tasks/delete
func HandleDeleteTask(repo *repository.TaskRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
			http.Error(w, "Invalid ID", http.StatusBadRequest)
			return
		}

		userID := r.Context().Value("user_id").(int)
		if err := repo.Delete(r.Context(), userID, input.ID); err != nil {
			http.Error(w, "Delete failed", http.StatusNotFound)
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
