package receipt

import (
	"fmt"
	"time"
	"unicode/utf8"

	"github.com/gmr458/receipt-processor-go/ecma262"
	"github.com/gmr458/receipt-processor-go/money"
	"github.com/gmr458/receipt-processor-go/timefmt"
	"github.com/gmr458/receipt-processor-go/validator"
)

type ReceiptDTO struct {
	Retailer     string      `json:"retailer"`
	PurchaseDate string      `json:"purchaseDate"`
	PurchaseTime string      `json:"purchaseTime"`
	Total        money.Money `json:"total"`
	Items        []ItemDTO   `json:"items"`
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
	retailerLen := utf8.RuneCountInString(dto.Retailer)

	v.Check(dto.Retailer != "", key, "retailer cannot be empty")
	v.Check(retailerLen <= maxLen, key, fmt.Sprintf("retailer max length is %d characters", maxLen))
	v.Check(ecma262.Only(dto.Retailer, isRetailerRune), key,
		"retailer can only contain letters, numbers, underscores, spaces, ampersands, and dashes",
	)
}

func (dto ReceiptDTO) ValidatePurchaseDate(v *validator.Validator) {
	const key = "purchaseDate"
	_, err := time.Parse(time.DateOnly, dto.PurchaseDate)
	v.Check(err == nil, key, "invalid format, it should be YYYY-MM-DD")
}

func (dto ReceiptDTO) ValidatePurchaseTime(v *validator.Validator) {
	const key = "purchaseTime"
	_, err := time.Parse(timefmt.TimeHourMinuteOnly, dto.PurchaseTime)
	v.Check(err == nil, key, "invalid format, it should be hh:mm")
}

func (dto ReceiptDTO) ValidateTotal(v *validator.Validator) {
	const key = "total"

	v.Check(dto.Total > 0, key, "the total must be greater than 0.0")
}

func (dto ReceiptDTO) ValidateItems(v *validator.Validator) {
	const key = "items"
	const maxLenShortDesc = 100

	v.Check(dto.Items != nil, key, "items cannot be null")
	v.Check(len(dto.Items) != 0, key, "items cannot be empty")

	for _, item := range dto.Items {
		lenShortDesc := utf8.RuneCountInString(item.ShortDescription)
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
			item.Price > 0,
			key,
			"there is one or more items that have a price of zero or less",
		)
		v.Check(ecma262.Only(item.ShortDescription, isShortDescriptionRune), key,
			"there is one or more items whose shortDescription contains characters other than letters, numbers, underscores, spaces, and dashes",
		)
	}
}

func (dto ReceiptDTO) ValidateTotalEqualItemsTotal(v *validator.Validator) {
	const key = "total"

	prices := make([]money.Money, len(dto.Items))
	for i, item := range dto.Items {
		prices[i] = item.Price
	}
	itemsTotal, sumErr := money.Sum(prices...)

	if sumErr != nil {
		// The items sum outside int64 cents, so the total can never equal
		// it; money.Money always errors instead of wrapping.
		v.Check(false, key, "the sum of all items price exceeds the representable range")
		return
	}

	message := fmt.Sprintf(
		"total field should be equal to the sum of all items price, total=%s != itemsTotal=%s",
		dto.Total, itemsTotal,
	)
	v.Check(dto.Total == itemsTotal, key, message)
}
