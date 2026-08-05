package receipt

import (
	"context"
	"math"
	"strings"
	"time"
)

type Receipt struct {
	ID           string    `json:"id"`
	Retailer     string    `json:"retailer"`
	PurchaseDate time.Time `json:"purchaseDate"`
	PurchaseTime time.Time `json:"purchaseTime"`
	Total        float64   `json:"total"`
	Items        []Item    `json:"items"`
}

type ReceiptRepository interface {
	Find(ctx context.Context, filters Filters) (PaginatedReceipts, error)
	FindById(ctx context.Context, id string) (*Receipt, error)
	Create(ctx context.Context, receipt *Receipt) error
}

type ReceiptCache interface {
	SetPaginatedReceipts(ctx context.Context, key string, paginatedReceipts PaginatedReceipts, exp time.Duration) error
	GetPaginatedReceipts(ctx context.Context, key string) (PaginatedReceipts, error)
	GetPointsById(ctx context.Context, id string) (int, error)
	SetPointsById(ctx context.Context, id string, points int, exp time.Duration) error
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
	if hasZeroDecimal(r.Total) {
		return 50
	}

	return 0
}

func (r Receipt) GetPointsTotalIsMultipleOf(f float64) int {
	if xIsMultipleOfy(r.Total, f) {
		return 25
	}

	return 0
}

func (r Receipt) GetPointsForEveryNItems(n int) int {
	points := 5
	return (len(r.Items) / n) * points
}

func (r Receipt) GetPointsItemsDescription() int {
	points := 0

	for _, item := range r.Items {
		trimmedLen := len(strings.TrimSpace(item.ShortDescription))
		if xIsMultipleOfy(float64(trimmedLen), 3.0) {
			p := int(math.Ceil(item.Price * 0.2))
			points += p
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
	points += r.GetPointsTotalIsMultipleOf(0.25)
	points += r.GetPointsForEveryNItems(2)
	points += r.GetPointsPurchaseDayIsOdd()
	points += r.GetPointsItemsDescription()
	points += r.GetPointsTimeOfPurchase()

	return points
}
