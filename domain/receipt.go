package domain

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

// GetPointsRetailerName returns one point for every alphanumeric character in the retailer name.
func (r Receipt) GetPointsRetailerName() int {
	points := 0

	for _, char := range r.Retailer {
		if isAlphanumeric(char) {
			points += 1
		}
	}

	return points
}

// GetPointsRoundDollar returns 50 points if the total is a round dollar amount with no cents.
func (r Receipt) GetPointsRoundDollar() int {
	if hasZeroDecimal(r.Total) {
		return 50
	}

	return 0
}

// GetPointsTotalIsMultipleOf returns 25 points if the total is a multiple of a given number.
func (r Receipt) GetPointsTotalIsMultipleOf(f float64) int {
	if xIsMultipleOfy(r.Total, f) {
		return 25
	}

	return 0
}

// GetPointsForEveryNItems returns 5 points for every n items on the receipt.
func (r Receipt) GetPointsForEveryNItems(n int) int {
	points := 5
	return (len(r.Items) / n) * points
}

// GetPointsItemsDescription returns a specific amount of points if the trimmed
// length of an item description is a multiple of 3, then it multiplies the price by 0.2
// and round up to the nearest integer, the result is the number of points earned.
// This is applied to every item.
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

// GetPointsPurchaseDayIsOdd returns 6 points if the day in the purchase date is odd.
func (r Receipt) GetPointsPurchaseDayIsOdd() int {
	day := r.PurchaseDate.Day()
	if isOdd(day) {
		return 6
	}

	return 0
}

// GetPointsTimeOfPurchase returns 10 points if the time of purchase is after 2:00pm and before 4:00pm.
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
