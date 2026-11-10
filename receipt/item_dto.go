package receipt

import "github.com/gmr458/receipt-processor-go/money"

type ItemDTO struct {
	ShortDescription string      `json:"shortDescription"`
	Price            money.Money `json:"price"`
}
