package main

import (
	"context"
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

/*
curl -X POST http://localhost:4000/feedback/create \
-H "Content-Type: application/json" \

	-d '{
	  "name": "Wasani",
	  "email": "wasani@example.com",
	  "subject": "Test Subject",
	  "message": "This is a test message"
	}'
*/
func (app *application) createFeedback(w http.ResponseWriter, r *http.Request) {
	var feedback Feedback

	err := app.readJSON(w, r, &feedback)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": err.Error(),
		})
		return
	}

	if feedback.Name == "" || feedback.Email == "" || feedback.Subject == "" || feedback.Message == "" {
		app.writeJSON(w, http.StatusBadRequest, map[string]string{
			"error": "Full name, email, subject and message are required",
		})
		return
	}

	query := `INSERT INTO feedback (name, email, subject, message)
	          VALUES ($1, $2, $3, $4) RETURNING feedback_id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, feedback.Name, feedback.Email, feedback.Subject, feedback.Message).Scan(&feedback.ID,
		&feedback.CreatedAt)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, map[string]string{
			"error": err.Error(),
		})
		return
	}

	err = app.writeJSON(w, http.StatusCreated, feedback)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/* curl "http://localhost:4000/feedback?id=6", curl "http://localhost:4000/feedback"*/
func (app *application) getFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	// Get single feedback
	if id != "" {
		var f Feedback

		query := `SELECT feedback_id, name, email, subject, message, created_at 
		          FROM feedback WHERE feedback_id = $1`

		err := app.db.QueryRowContext(ctx, query, id).Scan(
			&f.ID, &f.Name, &f.Email, &f.Subject, &f.Message, &f.CreatedAt,
		)

		if err != nil {
			app.writeJSON(w, 404, map[string]string{"error": "Feedback not found"})
			return
		}

		app.writeJSON(w, 200, f)
		return
	}

	//Get all feedback
	rows, err := app.db.QueryContext(ctx, `
		SELECT feedback_id, name, email, subject, message, created_at FROM feedback
	`)
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	defer rows.Close()

	var feedbacks []Feedback

	for rows.Next() {
		var f Feedback
		rows.Scan(&f.ID, &f.Name, &f.Email, &f.Subject, &f.Message, &f.CreatedAt)
		if err != nil {
			app.writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		feedbacks = append(feedbacks, f)
	}

	app.writeJSON(w, 200, feedbacks)
}

/*
curl -X PUT "http://localhost:4000/feedback/update?feedback_id=7" \

	-H "Content-Type: application/json" \
		-d '{
		  "name": "Wasani Updated",
		  "email": "wasani_new@example.com",
		  "subject": "Updated Subject",
		  "message": "Updated message"
		}'
*/
func (app *application) updateFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("feedback_id")
	if id == "" {
		app.writeJSON(w, 400, map[string]string{"error": "Missing feedback ID"})
		return
	}

	var feedback Feedback
	err := app.readJSON(w, r, &feedback)
	if err != nil {
		app.writeJSON(w, 400, map[string]string{"error": err.Error()})
		return
	}

	query := `UPDATE feedback SET name = $1, email = $2, subject = $3, message = $4
			  WHERE feedback_id = $5`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, feedback.Name, feedback.Email, feedback.Subject, feedback.Message, id)

	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		app.writeJSON(w, 404, map[string]string{"error": "Feedback not found"})
		return
	}

	app.writeJSON(w, 200, map[string]string{
		"message": "Your feedback was updated successfully",
	})
}

/* curl -X DELETE "http://localhost:4000/feedback/delete?feedback_id=6"*/
func (app *application) deleteFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("feedback_id")
	if id == "" {
		app.writeJSON(w, 400, map[string]string{"error": "Missing feedback id"})
		return
	}

	query := `DELETE FROM feedback WHERE feedback_id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, id)
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}

	if rowsAffected == 0 {
		app.writeJSON(w, 404, map[string]string{"error": "Feedback not found"})
		return
	}

	app.writeJSON(w, 200, map[string]string{
		"message": "Your feedback was deleted successfully",
	})
}

// curl http://localhost:4000/feedback/names
func (app *application) getNames(w http.ResponseWriter, r *http.Request) {
	query := `SELECT name FROM feedback`

	rows, err := app.db.Query(query)
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var names []string

	for rows.Next() {
		var n string
		err := rows.Scan(&n)
		if err != nil {
			app.writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		names = append(names, n)
	}

	app.writeJSON(w, 200, names)
}

// curl http://localhost:4000/feedback/emails
func (app *application) getEmails(w http.ResponseWriter, r *http.Request) {
	query := `SELECT email FROM feedback`

	rows, err := app.db.Query(query)
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var emails []string

	for rows.Next() {
		var e string
		err := rows.Scan(&e)
		if err != nil {
			app.writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		emails = append(emails, e)
	}

	app.writeJSON(w, 200, emails)
}

// curl http://localhost:4000/feedback/subjects
func (app *application) getSubjects(w http.ResponseWriter, r *http.Request) {
	query := `SELECT subject FROM feedback`

	rows, err := app.db.Query(query)
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var subjects []string

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			app.writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		subjects = append(subjects, s)
	}

	app.writeJSON(w, 200, subjects)
}

// //curl http://localhost:4000/feedback/messages
func (app *application) getMessages(w http.ResponseWriter, r *http.Request) {
	query := `SELECT message FROM feedback`

	rows, err := app.db.Query(query)
	if err != nil {
		app.writeJSON(w, 500, map[string]string{"error": err.Error()})
		return
	}
	defer rows.Close()

	var messages []string

	for rows.Next() {
		var m string
		err := rows.Scan(&m)
		if err != nil {
			app.writeJSON(w, 500, map[string]string{"error": err.Error()})
			return
		}
		messages = append(messages, m)
	}

	app.writeJSON(w, 200, messages)
}
