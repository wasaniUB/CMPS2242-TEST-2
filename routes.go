package main

import "net/http"

func (app *application) routes() http.Handler {
	mux := http.NewServeMux()

	// Feedback routes
	mux.HandleFunc("/feedback", app.getFeedback)
	mux.HandleFunc("/feedback/create", app.createFeedback)
	mux.HandleFunc("/feedback/update", app.updateFeedback)
	mux.HandleFunc("/feedback/delete", app.deleteFeedback)
	mux.HandleFunc("/feedback/names", app.getNames)
	mux.HandleFunc("/feedback/emails", app.getEmails)
	mux.HandleFunc("/feedback/subjects", app.getSubjects)
	mux.HandleFunc("/feedback/messages", app.getMessages)

	fileServer := http.FileServer(http.Dir("./Frontend"))

	mux.Handle("/", fileServer)

	return mux

}
