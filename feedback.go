package main

import (
	"context"
	"database/sql"
	"net/http"
	"time"
)

type Feedback struct {
	ID        int       `json:"feedback_id"`
	Name      string    `json:"name"`
	Email     string    `json:"email"`
	Subject   string    `json:"subject"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

/*
	curl -X POST http://localhost:4000/feedback/create \-H "Content-Type: application/json" \-d '{
				  "name": "BIGWAS",
				  "email": "bigwas@example.com",
				  "subject": "bigs",
				  "message": "ola"
				}'
*/
func (app *application) createFeedback(w http.ResponseWriter, r *http.Request) {
	var feedback Feedback

	err := app.readJSON(w, r, &feedback)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error": err.Error(),
		}, nil)
		return
	}

	if feedback.Name == "" || feedback.Email == "" || feedback.Subject == "" || feedback.Message == "" {
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error": "Full name, email, subject and message are required",
		}, nil)
		return
	}

	query := `INSERT INTO feedback (name, email, subject, message)
	          VALUES ($1, $2, $3, $4) RETURNING feedback_id, created_at`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	err = app.db.QueryRowContext(ctx, query, feedback.Name, feedback.Email, feedback.Subject, feedback.Message).Scan(&feedback.ID,
		&feedback.CreatedAt)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{
			"error": err.Error(),
		}, nil)
		return
	}

	err = app.writeJSON(w, http.StatusCreated, envelope{
		"feedback": feedback,
	}, nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

/* curl "http://localhost:4000/feedback?id=6", curl "http://localhost:4000/feedback"*/
func (app *application) getFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	//get feedback by id
	if id != "" {
		var f Feedback

		query := `SELECT feedback_id, name, email, subject, message, created_at 
		          FROM feedback WHERE feedback_id = $1`

		err := app.db.QueryRowContext(ctx, query, id).Scan(
			&f.ID, &f.Name, &f.Email, &f.Subject, &f.Message, &f.CreatedAt,
		)

		if err != nil {
			if err == sql.ErrNoRows {
				app.writeJSON(w, http.StatusNotFound, envelope{
					"error": "Feedback not found",
				}, nil)
				return
			}

			app.writeJSON(w, http.StatusInternalServerError, envelope{
				"error": err.Error(),
			}, nil)
			return
		}

		app.writeJSON(w, http.StatusOK, envelope{
			"feedback": f,
		}, nil)
		return
	}

	//get all feedbacks
	rows, err := app.db.QueryContext(ctx, `
		SELECT feedback_id, name, email, subject, message, created_at FROM feedback
	`)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{
			"error": err.Error(),
		}, nil)
		return
	}
	defer rows.Close()

	var feedbacks []Feedback

	for rows.Next() {
		var f Feedback

		err := rows.Scan(&f.ID, &f.Name, &f.Email, &f.Subject, &f.Message, &f.CreatedAt)
		if err != nil {
			app.writeJSON(w, http.StatusInternalServerError, envelope{
				"error": err.Error(),
			}, nil)
			return
		}

		feedbacks = append(feedbacks, f)
	}

	if err = rows.Err(); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{
			"error": err.Error(),
		}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{
		"feedback": feedbacks,
	}, nil)
}

/*
curl -X PUT "http://localhost:4000/feedback/update?feedback_id=9" \
-H "Content-Type: application/json" \
-d '{"name": "Wasani Updated", "email": "wasani_new@example.com", "subject": "Updated Subject", "message": "Updated message"}'
*/
func (app *application) updateFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("feedback_id")

	if id == "" {
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error": "Missing feedback id",
		}, nil)
		return
	}

	var feedback Feedback
	err := app.readJSON(w, r, &feedback)
	if err != nil {
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error": err.Error(),
		}, nil)
		return
	}

	if feedback.Name == "" || feedback.Email == "" || feedback.Subject == "" || feedback.Message == "" {
		app.writeJSON(w, http.StatusBadRequest, envelope{
			"error": "All fields are required",
		}, nil)
		return
	}

	query := `UPDATE feedback SET name = $1, email = $2, subject = $3, message = $4
			  WHERE feedback_id = $5`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, feedback.Name, feedback.Email, feedback.Subject, feedback.Message, id)

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{
			"error": err.Error(),
		}, nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{
			"error": err.Error(),
		}, nil)
		return
	}

	if rowsAffected == 0 {
		app.writeJSON(w, http.StatusNotFound, envelope{
			"error": "Feedback not found",
		}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{
		"message": "Your feedback was updated successfully",
	}, nil)
}

/* curl -X DELETE "http://localhost:4000/feedback/delete?feedback_id=12"*/
func (app *application) deleteFeedback(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("feedback_id")
	if id == "" {
		app.writeJSON(w, http.StatusBadRequest, envelope{"error": "Missing feedback id"}, nil)
		return
	}

	query := `DELETE FROM feedback WHERE feedback_id = $1`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()

	result, err := app.db.ExecContext(ctx, query, id)
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}

	if rowsAffected == 0 {
		app.writeJSON(w, http.StatusNotFound, envelope{"error": "Feedback not found"}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{
		"message": "Your feedback was deleted successfully",
	}, nil)
}

// curl http://localhost:4000/feedback/names
func (app *application) getNames(w http.ResponseWriter, r *http.Request) {
	query := `SELECT name FROM feedback`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx, query)

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}
	defer rows.Close()

	var names []string

	for rows.Next() {
		var n string
		err := rows.Scan(&n)
		if err != nil {
			app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
			return
		}
		names = append(names, n)
	}

	if err = rows.Err(); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"names": names}, nil)
}

//

func (app *application) getEmails(w http.ResponseWriter, r *http.Request) {
	query := `SELECT email FROM feedback`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx, query)

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}
	defer rows.Close()

	var emails []string

	for rows.Next() {
		var e string
		err := rows.Scan(&e)
		if err != nil {
			app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
			return
		}
		emails = append(emails, e)
	}

	if err = rows.Err(); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"emails": emails}, nil)
}

// curl http://localhost:4000/feedback/subjects
func (app *application) getSubjects(w http.ResponseWriter, r *http.Request) {
	query := `SELECT subject FROM feedback`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx, query)

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}
	defer rows.Close()

	var subjects []string

	for rows.Next() {
		var s string
		err := rows.Scan(&s)
		if err != nil {
			app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
			return
		}
		subjects = append(subjects, s)
	}

	if err = rows.Err(); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"subjects": subjects}, nil)
}

// //curl http://localhost:4000/feedback/messages
func (app *application) getMessages(w http.ResponseWriter, r *http.Request) {
	query := `SELECT message FROM feedback`

	ctx, cancel := context.WithTimeout(r.Context(), 3*time.Second)
	defer cancel()
	rows, err := app.db.QueryContext(ctx, query)

	if err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}
	defer rows.Close()

	var messages []string

	for rows.Next() {
		var m string
		err := rows.Scan(&m)
		if err != nil {
			app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
			return
		}
		messages = append(messages, m)
	}

	if err = rows.Err(); err != nil {
		app.writeJSON(w, http.StatusInternalServerError, envelope{"error": err.Error()}, nil)
		return
	}

	app.writeJSON(w, http.StatusOK, envelope{"messages": messages}, nil)
}
