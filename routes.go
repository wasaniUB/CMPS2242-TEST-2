package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Faculties
	mux.HandleFunc("/feedback", app.getFeedback)
	mux.HandleFunc("/feedback/create", app.createFeedback)
	mux.HandleFunc("/feedback/update", app.updateFeedback)
	mux.HandleFunc("/feedback/delete", app.deleteFeedback)

	return mux

}
