package receipt

import "testing"

// baseValidDTO is a valid challenge examples body; pattern tests vary only
// the field under test so failures are attributable.
func baseValidDTO() ReceiptDTO {
	return ReceiptDTO{
		Retailer:     "Target",
		PurchaseDate: "2022-01-01",
		PurchaseTime: "13:01",
		Total:        649,
		Items:        []ItemDTO{{ShortDescription: "Mountain Dew 12PK", Price: 649}},
	}
}

// TestRetailerPattern enforces the api.yml retailer pattern "^[\\w\\s\\-&]+$"
// interpreted with ECMA-262 semantics (OpenAPI 3.0). Verified equivalent
// against JavaScript's regex engine.
func TestRetailerPattern(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"challenge example", "Target", true},
		{"ampersand allowed", "M&M Corner Market", true},
		{"underscore and dash", "Target_1-2", true},
		{"tab is whitespace", "Target\tStore", true},
		{"NBSP is whitespace for ECMA", "A\u00A0B", true},
		{"vertical tab is whitespace for ECMA", "A\vB", true},
		{"whitespace-only passes the pattern", "   ", true},
		{"not-ampersand rejected", "Target!", false},
		{"unicode letters rejected", "Café", false},
		{"emoji rejected", "🚀", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := baseValidDTO()
			dto.Retailer = tt.value

			isValid, _ := dto.IsValid()
			if isValid != tt.want {
				t.Errorf("IsValid with retailer %q = %v, want %v", tt.value, isValid, tt.want)
			}
		})
	}
}

// TestShortDescriptionPattern enforces the api.yml item pattern
// "^[\\w\\s\\-]+$": unlike the retailer, the ampersand is NOT allowed.
func TestShortDescriptionPattern(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"challenge example", "Mountain Dew 12PK", true},
		{"dash and digits", "Pepsi - 12-oz", true},
		{"underscore", "Dasani_1", true},
		{"NBSP is whitespace for ECMA", "A\u00A0B", true},
		{"ampersand not allowed", "Milk & Eggs", false},
		{"unicode letters rejected", "Café", false},
		{"emoji rejected", "🚀", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			dto := baseValidDTO()
			dto.Items = []ItemDTO{{ShortDescription: tt.value, Price: 649}}

			isValid, errs := dto.IsValid()
			if isValid != tt.want {
				t.Errorf("IsValid with shortDescription %q = %v (%v), want %v", tt.value, isValid, errs, tt.want)
			}
		})
	}
}
