package main

import (
	"context"
	"encoding/json"
	"net/http"
	"time"
)

type Feedback struct {
	ID        int    `json:"feedback_id"`
	Name      string `json:"name"`
	Email     string `json:"email"`
	Subject   string `json:"subject"`
	Message   string `json:"message"`
	CreatedAt string `json:"created_at"`
}

func (app *application) createFeedback(w http.ResponseWriter, r *http.Request) {
	var feedback Feedback

	err := json.NewDecoder(r.Body).Decode(&feedback)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if feedback.Name == "" || feedback.Email == "" || feedback.Subject == "" || feedback.Message == "" {
		http.Error(w, "Full name, email, subject and message are required", http.StatusBadRequest)
		return
	}

	query := `INSERT INTO feedback (name, email, subject, message)
	          VALUES ($1, $2, $3, $4) RETURNING feedback_id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, feedback.Name, feedback.Email, feedback.Subject, feedback.Message).Scan(&feedback.ID,
		&feedback.CreatedAt)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(feedback)
}

func (app *application) getFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id", http.StatusBadRequest)
		return
	}

	query := `SELECT feedback_id, name, email, subject, message, created_at 
	          FROM feedback WHERE feedback_id = $1`

	var f Feedback

	err := app.db.QueryRow(query, id).Scan(&f.ID, &f.Name, &f.Email, &f.Subject, &f.Message, &f.CreatedAt)

	if err != nil {
		http.Error(w, "Feedback not found", http.StatusNotFound)
		return
	}

	json.NewEncoder(w).Encode(f)
}

func (app *application) updateFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("feedback_id")
	if id == "" {
		http.Error(w, "Missing feedback ID\n", http.StatusBadRequest)
		return
	}

	var feedback Feedback
	err := json.NewDecoder(r.Body).Decode(&feedback)
	if err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	query := `UPDATE feedback SET name = $1, email = $2, subject = $3, message = $4
	          WHERE feedback_id = $5`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, feedback.Name, feedback.Email, feedback.Subject, feedback.Message, id)

	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Feedback not found", 404)
		return
	}

	json.NewEncoder(w).Encode(map[string]string{
		"message": "Your feedback was updated successfully",
	})
}

func (app *application) deleteFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("feedback_id")
	if id == "" {
		http.Error(w, "Missing feedback id\n", http.StatusBadRequest)
		return
	}

	query := `DELETE FROM feedback WHERE feedback_id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, id)
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if rowsAffected == 0 {
		http.Error(w, "Feedback not found", 404)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Your feedback was deleted successfully\n"))
}
