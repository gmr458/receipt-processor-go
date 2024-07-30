package main

import (
	"net/http"
	"time"

	"github.com/google/uuid"

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

	receipt := &domain.Receipt{
		ID:       uuid.New().String(),
		Retailer: input.Retailer,
		Total:    input.Total,
		Items:    []domain.Item{},
	}
	receipt.PurchaseDate, _ = time.Parse("2006-01-02", input.PurchaseDate)
	receipt.PurchaseTime, _ = time.Parse("15:04", input.PurchaseTime)

	for _, itemDto := range input.Items {
		item := domain.Item{
			ID:               uuid.New().String(),
			ShortDescription: itemDto.ShortDescription,
			Price:            itemDto.Price,
		}
		receipt.Items = append(receipt.Items, item)
	}

	err = app.service.Receipt.Create(r.Context(), receipt)
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

	receipt, err := app.service.Receipt.FindById(r.Context(), id)
	if err != nil {
		app.errorResponse(w, r, err)
		return
	}

	points := receipt.CalculateTotalPoints()

	app.sendJSON(w, http.StatusOK, envelope{
		"points": points,
	}, nil)
}
