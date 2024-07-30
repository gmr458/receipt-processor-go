package domain

type Item struct {
	ID               string  `json:"id"`
	ShortDescription string  `json:"shortDescription"`
	Price            float64 `json:"price"`
	ReceiptID        string  `json:"receiptID"`
}
