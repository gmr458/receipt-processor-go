package receipt

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/gmr458/receipt-processor-go/money"
)

// The two canonical examples from the official receipt-processor challenge
// README, with the published expected outputs.
const (
	readmeTargetExample = `{
  "retailer": "Target",
  "purchaseDate": "2022-01-01",
  "purchaseTime": "13:01",
  "items": [
    {"shortDescription": "Mountain Dew 12PK", "price": "6.49"},
    {"shortDescription": "Emils Cheese Pizza", "price": "12.25"},
    {"shortDescription": "Knorr Creamy Chicken", "price": "1.26"},
    {"shortDescription": "Doritos Nacho Cheese", "price": "3.35"},
    {"shortDescription": "   Klarbrunn 12-PK 12 FL OZ  ", "price": "12.00"}
  ],
  "total": "35.35"
}`

	readmeMMExample = `{
  "retailer": "M&M Corner Market",
  "purchaseDate": "2022-03-20",
  "purchaseTime": "14:33",
  "items": [
    {"shortDescription": "Gatorade", "price": "2.25"},
    {"shortDescription": "Gatorade", "price": "2.25"},
    {"shortDescription": "Gatorade", "price": "2.25"},
    {"shortDescription": "Gatorade", "price": "2.25"}
  ],
  "total": "9.00"
}`
)

// exampleRepo stores the processed receipt so tests can inspect it.
type exampleRepo struct {
	saved *Receipt
}

func (r *exampleRepo) Find(_ context.Context, _ Filters) (PaginatedReceipts, error) {
	return PaginatedReceipts{}, nil
}

func (r *exampleRepo) FindById(_ context.Context, _ string) (*Receipt, error) {
	return nil, nil
}

func (r *exampleRepo) Create(_ context.Context, rec *Receipt) error {
	r.saved = rec
	return nil
}

// TestChallengeExampleReceipts runs the official README examples through the
// full decode → validate → process → points pipeline and asserts the
// published outputs. The simple/morning examples come from the challenge's
// examples directory; their expected points are derived from the rules.
func TestChallengeExampleReceipts(t *testing.T) {
	cases := []struct {
		name       string
		body       string
		wantTotal  money.Money
		wantPoints int
	}{
		{
			"README example: Target (Total Points: 28)",
			readmeTargetExample,
			3535,
			28,
		},
		{
			"README example: M&M Corner Market (Total Points: 109)",
			readmeMMExample,
			900,
			109,
		},
		{
			"examples/simple-receipt.json (retailer 6 + total multiple of 0.25)",
			`{"retailer": "Target", "purchaseDate": "2022-01-02", "purchaseTime": "13:13", "total": "1.25", "items": [{"shortDescription": "Pepsi - 12-oz", "price": "1.25"}]}`,
			125,
			31,
		},
		{
			"examples/morning-receipt.json (retailer 9 + one item pair + Dasani desc 6 chars/3 → 1)",
			`{"retailer": "Walgreens", "purchaseDate": "2022-01-02", "purchaseTime": "08:13", "total": "2.65", "items": [{"shortDescription": "Pepsi - 12-oz", "price": "1.25"}, {"shortDescription": "Dasani", "price": "1.40"}]}`,
			265,
			15,
		},
	}

	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			var dto ReceiptDTO
			if err := json.Unmarshal([]byte(tt.body), &dto); err != nil {
				t.Fatalf("decode: %v", err)
			}
			if valid, errs := dto.IsValid(); !valid {
				t.Fatalf("valid example rejected: %v", errs)
			}

			repo := &exampleRepo{}
			svc := NewService(repo)
			if _, err := svc.Process(context.Background(), dto); err != nil {
				t.Fatalf("process: %v", err)
			}

			rec := repo.saved
			if rec.Total != tt.wantTotal {
				t.Errorf("total = %d, want %d", rec.Total, tt.wantTotal)
			}
			if got := rec.CalculateTotalPoints(); got != tt.wantPoints {
				t.Errorf("points = %d, want %d", got, tt.wantPoints)
			}
		})
	}
}
