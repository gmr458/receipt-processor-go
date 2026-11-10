package receipt

import "github.com/gmr458/receipt-processor-go/money"

type Item struct {
	ID               string      `json:"id"`
	ShortDescription string      `json:"shortDescription"`
	Price            money.Money `json:"price"`
	ReceiptID        string      `json:"receiptID"`
}
