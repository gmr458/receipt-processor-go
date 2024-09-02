package repository

import (
	"github.com/gmr458/receipt-processor/domain"
	"github.com/gmr458/receipt-processor/sqlite"
)

type Repository struct {
	Receipt domain.ReceiptRepository
}

func New(sqliteConn *sqlite.Conn) Repository {
	return Repository{
		Receipt: ReceiptRepository{sqliteConn},
	}
}
