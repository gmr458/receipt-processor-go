package main

import (
	"net/http"

	"github.com/gmr458/receipt-processor-go/receipt"
)

func (app *app) handlerProcessReceipts(w http.ResponseWriter, r *http.Request) {
	var input receipt.ReceiptDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	receipt, err := app.receiptService.Process(r.Context(), input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"id": receipt.ID,
	}, nil)
}

func (app *app) handlerGetPoints(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if id == "" {
		app.badRequest(w, "Invalid path value", map[string]string{
			"id": "id cannot be an empty string",
		})
		return
	}

	points, err := app.receiptService.GetPointsById(r.Context(), id)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"points": points,
	}, nil)
}

func (app *app) handlerGetReceipts(w http.ResponseWriter, r *http.Request) {
	filters, err := NewFiltersFromURL(
		r.URL.Query(),
		"purchase_date",
		"id",
		"-id",
		"retailer",
		"-retailer",
		"purchase_date",
		"-purchase_date",
		"total",
		"-total",
	)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	paginatedReceipts, err := app.receiptService.GetReceipts(r.Context(), filters)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, paginatedReceipts, nil)
}
