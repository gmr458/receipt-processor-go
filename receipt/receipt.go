package receipt

import (
	"context"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/gmr458/receipt-processor-go/money"
)

type Receipt struct {
	ID           string      `json:"id"`
	Retailer     string      `json:"retailer"`
	PurchaseDate time.Time   `json:"purchaseDate"`
	PurchaseTime time.Time   `json:"purchaseTime"`
	Total        money.Money `json:"total"`
	Items        []Item      `json:"items"`
}

type ReceiptRepository interface {
	Find(ctx context.Context, filters Filters) (PaginatedReceipts, error)
	FindById(ctx context.Context, id string) (*Receipt, error)
	Create(ctx context.Context, receipt *Receipt) error
}

type PaginatedReceipts struct {
	Receipts []Receipt `json:"receipts"`
	Metadata *Metadata `json:"metadata"`
}

func (r Receipt) GetPointsRetailerName() int {
	points := 0

	for _, char := range r.Retailer {
		if isAlphanumeric(char) {
			points += 1
		}
	}

	return points
}

func (r Receipt) GetPointsRoundDollar() int {
	if r.Total%100 == 0 {
		return 50
	}

	return 0
}

func (r Receipt) GetPointsTotalIsMultipleOf(m money.Money) int {
	if r.Total%m == 0 {
		return 25
	}

	return 0
}

func (r Receipt) GetPointsForEveryNItems(n int) int {
	if n <= 0 {
		return 0
	}

	return (len(r.Items) / n) * 5
}

func (r Receipt) GetPointsItemsDescription() int {
	points := 0

	for _, item := range r.Items {
		trimmedLen := utf8.RuneCountInString(strings.TrimSpace(item.ShortDescription))
		if trimmedLen%3 == 0 {
			bonus, err := item.Price.MulRatio(1, 500, money.Up)
			if err != nil {
				continue
			}
			points += int(bonus)
		}
	}

	return points
}

func (r Receipt) GetPointsPurchaseDayIsOdd() int {
	day := r.PurchaseDate.Day()
	if isOdd(day) {
		return 6
	}

	return 0
}

func (r Receipt) GetPointsTimeOfPurchase() int {
	hours, mins, _ := r.PurchaseTime.Clock()
	if hours == 14 && mins > 0 {
		return 10
	}
	if hours == 15 {
		return 10
	}
	return 0
}

func (r *Receipt) CalculateTotalPoints() int {
	points := 0

	points += r.GetPointsRetailerName()
	points += r.GetPointsRoundDollar()
	points += r.GetPointsTotalIsMultipleOf(25)
	points += r.GetPointsForEveryNItems(2)
	points += r.GetPointsPurchaseDayIsOdd()
	points += r.GetPointsItemsDescription()
	points += r.GetPointsTimeOfPurchase()

	return points
}
