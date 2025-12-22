package domain

import (
	"testing"
	"time"
)

func TestCalculatePoints(t *testing.T) {
	tests := []struct {
		receipt                         Receipt
		expectedPointsTotal             int
		expectedPointsRetailerName      int
		expectedPointsRoundDollar       int
		expectedPointsTotalIsMultipleOf int
		expectedPointsForEveryTwoItems  int
		expectedPointsDateIsOdd         int
		expectedPointsDescription       int
		expectedPointsTimeOfPurchase    int
	}{
		{
			receipt: Receipt{
				Retailer:     "Target",
				PurchaseDate: time.Date(2022, time.January, 1, 13, 1, 0, 0, time.Local),
				PurchaseTime: time.Date(2022, 1, 1, 13, 1, 0, 0, time.Local),
				Items: []Item{
					{
						ShortDescription: "Mountain Dew 12PK",
						Price:            6.49,
					},
					{
						ShortDescription: "Emils Cheese Pizza",
						Price:            12.25,
					},
					{
						ShortDescription: "Knorr Creamy Chicken",
						Price:            1.26,
					},
					{
						ShortDescription: "Doritos Nacho Cheese",
						Price:            3.35,
					},
					{
						ShortDescription: "   Klarbrunn 12-PK 12 FL OZ  ",
						Price:            12.00,
					},
				},
				Total: 35.35,
			},
			expectedPointsTotal:             28,
			expectedPointsRetailerName:      6,
			expectedPointsRoundDollar:       0,
			expectedPointsTotalIsMultipleOf: 0,
			expectedPointsForEveryTwoItems:  10,
			expectedPointsDateIsOdd:         6,
			expectedPointsDescription:       6,
			expectedPointsTimeOfPurchase:    0,
		},
		{
			receipt: Receipt{
				Retailer:     "M&M Corner Market",
				PurchaseDate: time.Date(2022, time.March, 20, 14, 33, 0, 0, time.Local),
				PurchaseTime: time.Date(2022, time.March, 20, 14, 33, 0, 0, time.Local),
				Items: []Item{
					{
						ShortDescription: "Gatorade",
						Price:            2.25,
					},
					{
						ShortDescription: "Gatorade",
						Price:            2.25,
					},
					{
						ShortDescription: "Gatorade",
						Price:            2.25,
					},
					{
						ShortDescription: "Gatorade",
						Price:            2.25,
					},
				},
				Total: 9.00,
			},
			expectedPointsTotal:             109,
			expectedPointsRetailerName:      14,
			expectedPointsRoundDollar:       50,
			expectedPointsTotalIsMultipleOf: 25,
			expectedPointsForEveryTwoItems:  10,
			expectedPointsDateIsOdd:         0,
			expectedPointsDescription:       0,
			expectedPointsTimeOfPurchase:    10,
		},
		{
			receipt: Receipt{
				Retailer:     "Test Store",
				PurchaseDate: time.Date(2022, time.January, 1, 15, 0, 0, 0, time.Local),
				PurchaseTime: time.Date(2022, 1, 1, 15, 0, 0, 0, time.Local),
				Items: []Item{
					{
						ShortDescription: "Item",
						Price:            1.00,
					},
				},
				Total: 1.00,
			},
			expectedPointsTotal:             100,
			expectedPointsRetailerName:      9,
			expectedPointsRoundDollar:       50,
			expectedPointsTotalIsMultipleOf: 25,
			expectedPointsForEveryTwoItems:  0,
			expectedPointsDateIsOdd:         6,
			expectedPointsDescription:       0,
			expectedPointsTimeOfPurchase:    10,
		},
		{
			receipt: Receipt{
				Retailer:     "Test Store",
				PurchaseDate: time.Date(2022, time.January, 1, 14, 0, 0, 0, time.Local),
				PurchaseTime: time.Date(2022, 1, 1, 14, 0, 0, 0, time.Local),
				Items: []Item{
					{
						ShortDescription: "Item",
						Price:            1.00,
					},
				},
				Total: 1.00,
			},
			expectedPointsTotal:             90,
			expectedPointsRetailerName:      9,
			expectedPointsRoundDollar:       50,
			expectedPointsTotalIsMultipleOf: 25,
			expectedPointsForEveryTwoItems:  0,
			expectedPointsDateIsOdd:         6,
			expectedPointsDescription:       0,
			expectedPointsTimeOfPurchase:    0,
		},
		{
			receipt: Receipt{
				Retailer:     "Test Store",
				PurchaseDate: time.Date(2022, time.January, 1, 15, 59, 0, 0, time.Local),
				PurchaseTime: time.Date(2022, 1, 1, 15, 59, 0, 0, time.Local),
				Items: []Item{
					{
						ShortDescription: "Item",
						Price:            1.00,
					},
				},
				Total: 1.00,
			},
			expectedPointsTotal:             100,
			expectedPointsRetailerName:      9,
			expectedPointsRoundDollar:       50,
			expectedPointsTotalIsMultipleOf: 25,
			expectedPointsForEveryTwoItems:  0,
			expectedPointsDateIsOdd:         6,
			expectedPointsDescription:       0,
			expectedPointsTimeOfPurchase:    10,
		},
		{
			receipt: Receipt{
				Retailer:     "Test Store",
				PurchaseDate: time.Date(2022, time.January, 1, 16, 0, 0, 0, time.Local),
				PurchaseTime: time.Date(2022, 1, 1, 16, 0, 0, 0, time.Local),
				Items: []Item{
					{
						ShortDescription: "Item",
						Price:            1.00,
					},
				},
				Total: 1.00,
			},
			expectedPointsTotal:             90,
			expectedPointsRetailerName:      9,
			expectedPointsRoundDollar:       50,
			expectedPointsTotalIsMultipleOf: 25,
			expectedPointsForEveryTwoItems:  0,
			expectedPointsDateIsOdd:         6,
			expectedPointsDescription:       0,
			expectedPointsTimeOfPurchase:    0,
		},
	}

	for _, tt := range tests {
		totalPoints := tt.receipt.CalculateTotalPoints()
		if totalPoints != tt.expectedPointsTotal {
			t.Errorf(
				"expected total points to be %d. got %d",
				tt.expectedPointsTotal, totalPoints,
			)
		}

		retailerNamePoints := tt.receipt.GetPointsRetailerName()
		if retailerNamePoints != tt.expectedPointsRetailerName {
			t.Errorf(
				"expected retailer name points to be %d. got %d",
				tt.expectedPointsRetailerName, retailerNamePoints,
			)
		}

		roundDollarPoints := tt.receipt.GetPointsRoundDollar()
		if roundDollarPoints != tt.expectedPointsRoundDollar {
			t.Errorf(
				"expected round dollar points to be %d. got %d",
				tt.expectedPointsRoundDollar, roundDollarPoints,
			)
		}

		totalIsMultipleOfPoints := tt.receipt.GetPointsTotalIsMultipleOf(0.25)
		if totalIsMultipleOfPoints != tt.expectedPointsTotalIsMultipleOf {
			t.Errorf(
				"expected total is multiple of 0.25 points to be %d. got %d",
				tt.expectedPointsTotalIsMultipleOf, totalIsMultipleOfPoints,
			)
		}

		everyTwoItemsPoints := tt.receipt.GetPointsForEveryNItems(2)
		if everyTwoItemsPoints != tt.expectedPointsForEveryTwoItems {
			t.Errorf(
				"expected for every two items of points to be %d. got %d",
				tt.expectedPointsForEveryTwoItems, everyTwoItemsPoints,
			)
		}

		dateIsOddPoints := tt.receipt.GetPointsPurchaseDayIsOdd()
		if dateIsOddPoints != tt.expectedPointsDateIsOdd {
			t.Errorf(
				"expected date is odd points to be %d. got %d",
				tt.expectedPointsDateIsOdd, dateIsOddPoints,
			)
		}

		descriptionPoints := tt.receipt.GetPointsItemsDescription()
		if descriptionPoints != tt.expectedPointsDescription {
			t.Errorf(
				"expected descriptions points to be %d. got %d",
				tt.expectedPointsDescription, descriptionPoints,
			)
		}

		timeOfPurchasePoints := tt.receipt.GetPointsTimeOfPurchase()
		if timeOfPurchasePoints != tt.expectedPointsTimeOfPurchase {
			t.Errorf(
				"expected time of purchase points to be %d. got %d",
				tt.expectedPointsTimeOfPurchase, timeOfPurchasePoints,
			)
		}
	}
}
