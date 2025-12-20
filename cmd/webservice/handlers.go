package main

import (
	"net/http"

	"github.com/gmr458/receipt-processor/domain"
)

func (app *app) handlerProcessReceipts(w http.ResponseWriter, r *http.Request) {
	var input domain.ReceiptDTO

	err := app.readJSON(w, r, &input)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	receipt, err := app.service.Receipt.Process(r.Context(), input)
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

func (app *app) handlerGetReceipts(w http.ResponseWriter, r *http.Request) {
	queryValues := r.URL.Query()
	filters := domain.NewFilters(
		"id",
		"-id",
		"retailer",
		"-retailer",
		"purchase_date",
		"-purchase_date",
		"total",
		"-total",
	)
	filters.Page = getURLValuePositiveInt(queryValues, "page", 1)
	filters.Limit = getURLValuePositiveInt(queryValues, "limit", 10)
	filters.Sort = getURLValueStr(queryValues, filters.SortSafeList, "sort", "purchase_date")

	paginatedReceipts, err := app.service.Receipt.GetReceipts(r.Context(), filters)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	app.sendJSON(w, http.StatusOK, paginatedReceipts, nil)
}
