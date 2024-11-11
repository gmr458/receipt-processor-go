package main

import (
	"net/http"

	"github.com/go-fuego/fuego"

	"github.com/gmr458/receipt-processor/domain"
	"github.com/gmr458/receipt-processor/validator"
)

type Item struct {
	ShortDescription string  `json:"shortDescription"`
	Price            float64 `json:"price"`
}

type ReceiptCreate struct {
	Retailer     string  `json:"retailer"`
	PurchaseDate string  `json:"purchaseDate"`
	PurchaseTime string  `json:"purchaseTime"`
	Total        float64 `json:"total"`
	Items        []Item  `json:"items"`
}

type RespProcessReceipts struct {
	Id string `json:"id"`
}

func (app *app) handlerProcessReceipts(
	c *fuego.ContextWithBody[domain.ReceiptDTO],
) (RespProcessReceipts, error) {
	body, err := c.Body()
	if err != nil {
		return RespProcessReceipts{}, err
	}

	body.Validator = validator.New()
	body.Validate()
	if !body.Validator.Ok() {
		var errs []fuego.ErrorItem
		for k, v := range body.Validator.Errors {
			errs = append(errs, fuego.ErrorItem{
				Name:   k,
				Reason: v,
			})
		}

		return RespProcessReceipts{}, fuego.BadRequestError{
			Title:  http.StatusText(http.StatusBadRequest),
			Status: http.StatusBadRequest,
			Detail: "Invalid field/s",
			Errors: errs,
		}
	}

	receipt, err := app.service.Receipt.Process(c.Context(), &body)
	if err != nil {
		return RespProcessReceipts{}, err
	}

	c.SetStatus(http.StatusCreated)
	return RespProcessReceipts{Id: receipt.ID}, nil
}

type RespGetPoints struct {
	Points int `json:"points"`
}

func (app *app) handlerGetPoints(c *fuego.ContextNoBody) (RespGetPoints, error) {
	id := c.PathParam("id")
	if id == "" {
		return RespGetPoints{}, fuego.BadRequestError{
			Title:  http.StatusText(http.StatusBadRequest),
			Status: http.StatusBadRequest,
			Detail: "Invalid path value",
			Errors: []fuego.ErrorItem{
				{Name: "id", Reason: "id cannot be an empty string"},
			},
		}
	}

	points, err := app.service.Receipt.GetPointsById(c.Request().Context(), id)
	if err != nil {
		return RespGetPoints{}, err
	}

	return RespGetPoints{Points: points}, nil
}
