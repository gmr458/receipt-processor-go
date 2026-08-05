package sqlite

import (
	"github.com/gmr458/receipt-processor/receipt"
)

type Repository struct {
	Receipt receipt.ReceiptRepository
}

func NewRepository(conn *Conn) Repository {
	return Repository{
		Receipt: ReceiptRepository{conn},
	}
}
