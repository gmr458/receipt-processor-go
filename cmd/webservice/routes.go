package main

import "net/http"

func (app *app) setupRoutes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /receipts/process", app.handlerProcessReceipts)
	mux.HandleFunc("GET /receipts/{id}/points", app.handlerGetPoints)
	mux.HandleFunc("GET /receipts", app.handlerGetReceipts)

	return app.recoverPanic(app.cors(app.rateLimit(mux)))
}
