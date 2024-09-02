package main

import (
	"net/http"

	"github.com/gmr458/receipt-processor/domain"
	"github.com/gmr458/receipt-processor/validator"
)

func (app *app) handlerProcessReceipts(w http.ResponseWriter, r *http.Request) {
	input := domain.ReceiptDTO{Validator: validator.New()}

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	input.Validate()
	if !input.Validator.Ok() {
		app.badRequest(w, "Invalid field/s", input.Validator.Errors)
		return
	}

	receipt, err := app.service.Receipt.Process(r.Context(), &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusCreated, envelope{
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

	points, err := app.service.Receipt.GetPointsById(r.Context(), id)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, envelope{
		"points": points,
	}, nil)
}
