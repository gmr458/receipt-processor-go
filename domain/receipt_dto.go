package domain

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gmr458/receipt-processor/validator"
)

type ReceiptDTO struct {
	Validator *validator.Validator `json:"-"`

	Retailer     string    `json:"retailer"`
	PurchaseDate string    `json:"purchaseDate"`
	PurchaseTime string    `json:"purchaseTime"`
	Total        float64   `json:"total"`
	Items        []ItemDTO `json:"items"`
}

func NewReceiptDTO() *ReceiptDTO {
	return &ReceiptDTO{
		Validator: validator.New(),
	}
}

func (dto *ReceiptDTO) Validate() {
	dto.ValidateRetailer()
	dto.ValidatePurchaseDate()
	dto.ValidatePurchaseTime()
	dto.ValidateTotal()
	dto.ValidateItems()
	dto.ValidateTotalEqualItemsTotal()
}

func (dto *ReceiptDTO) ValidateRetailer() {
	const key = "retailer"
	const maxLen = 50
	retailerLen := len(dto.Retailer)

	dto.Validator.Check(dto.Retailer != "", key, "retailer cannot be empty")
	dto.Validator.Check(retailerLen <= maxLen, key, fmt.Sprintf("retailer max length is %d characters", maxLen))
}

func (dto *ReceiptDTO) ValidatePurchaseDate() {
	const key = "purchaseDate"
	_, err := time.Parse("2006-01-02", dto.PurchaseDate)
	dto.Validator.Check(err == nil, key, "invalid format, it should be YYYY-MM-DD")
}

func (dto *ReceiptDTO) ValidatePurchaseTime() {
	const key = "purchaseTime"
	_, err := time.Parse("15:04", dto.PurchaseTime)
	dto.Validator.Check(err == nil, key, "invalid format, it should be hh:mm")
}

func (dto *ReceiptDTO) ValidateTotal() {
	const key = "total"

	dto.Validator.Check(dto.Total > 0.0, key, "the total must be greater than 0.0")
}

func (dto *ReceiptDTO) ValidateItems() {
	const key = "items"
	const maxLenShortDesc = 100

	for _, item := range dto.Items {
		lenShortDesc := len(item.ShortDescription)
		dto.Validator.Check(item.ShortDescription != "",
			key,
			"there are one or more items that have an empty shortDescription",
		)
		dto.Validator.Check(
			lenShortDesc <= maxLenShortDesc,
			key,
			fmt.Sprintf(
				"the length of shortDescription must be a maximum of %d characters",
				maxLenShortDesc,
			),
		)
		dto.Validator.Check(
			item.Price > 0.0,
			key,
			"there is one or more items that have a price of zero or less",
		)
	}
}

func (dto *ReceiptDTO) ValidateTotalEqualItemsTotal() {
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
	dto.Validator.Check(totalFormatted == itemsTotalFormatted, key, message)
}
