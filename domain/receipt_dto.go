package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gmr458/receipt-processor/validator"
)

type ReceiptDTO struct {
	Retailer     string    `json:"retailer"`
	PurchaseDate string    `json:"purchaseDate"`
	PurchaseTime string    `json:"purchaseTime"`
	Total        float64   `json:"total"`
	Items        []ItemDTO `json:"items"`
}

func (dto ReceiptDTO) IsValid() (bool, map[string]string) {
	v := validator.New()

	dto.ValidateRetailer(v)
	dto.ValidatePurchaseDate(v)
	dto.ValidatePurchaseTime(v)
	dto.ValidateTotal(v)
	dto.ValidateItems(v)
	dto.ValidateTotalEqualItemsTotal(v)

	return v.Ok(), v.Errors
}

func (dto ReceiptDTO) ValidateRetailer(v *validator.Validator) {
	const key = "retailer"
	const maxLen = 50
	retailerLen := len(dto.Retailer)

	v.Check(dto.Retailer != "", key, "retailer cannot be empty")
	v.Check(retailerLen <= maxLen, key, fmt.Sprintf("retailer max length is %d characters", maxLen))
}

func (dto ReceiptDTO) ValidatePurchaseDate(v *validator.Validator) {
	const key = "purchaseDate"
	_, err := time.Parse("2006-01-02", dto.PurchaseDate)
	v.Check(err == nil, key, "invalid format, it should be YYYY-MM-DD")
}

func (dto ReceiptDTO) ValidatePurchaseTime(v *validator.Validator) {
	const key = "purchaseTime"
	_, err := time.Parse("15:04", dto.PurchaseTime)
	v.Check(err == nil, key, "invalid format, it should be hh:mm")
}

func (dto ReceiptDTO) ValidateTotal(v *validator.Validator) {
	const key = "total"

	v.Check(dto.Total > 0.0, key, "the total must be greater than 0.0")
}

func (dto ReceiptDTO) ValidateItems(v *validator.Validator) {
	const key = "items"
	const maxLenShortDesc = 100

	v.Check(dto.Items != nil, key, "items cannot be null")
	v.Check(len(dto.Items) != 0, key, "items cannot be empty")

	for _, item := range dto.Items {
		lenShortDesc := len(item.ShortDescription)
		v.Check(item.ShortDescription != "",
			key,
			"there are one or more items that have an empty shortDescription",
		)
		v.Check(
			lenShortDesc <= maxLenShortDesc,
			key,
			fmt.Sprintf(
				"the length of shortDescription must be a maximum of %d characters",
				maxLenShortDesc,
			),
		)
		v.Check(
			item.Price > 0.0,
			key,
			"there is one or more items that have a price of zero or less",
		)
	}
}

func (dto ReceiptDTO) ValidateTotalEqualItemsTotal(v *validator.Validator) {
	const key = "total"

	itemsTotal := 0.0
	for _, item := range dto.Items {
		itemsTotal += item.Price
	}

	totalFormatted := strconv.FormatFloat(dto.Total, 'f', 2, 64)
	itemsTotalFormatted := strconv.FormatFloat(itemsTotal, 'f', 2, 64)

	message := fmt.Sprintf(
		"total field should be equal to the sum of all items price, total=%.2f != itemsTotal=%.2f",
		dto.Total,
		itemsTotal,
	)
	v.Check(totalFormatted == itemsTotalFormatted, key, message)
}
