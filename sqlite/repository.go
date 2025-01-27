package sqlite

import (
	"github.com/gmr458/receipt-processor/domain"
)

type Repository struct {
	Receipt domain.ReceiptRepository
}

func NewRepository(conn *Conn) Repository {
	return Repository{
		Receipt: ReceiptRepository{conn},
	}
}
