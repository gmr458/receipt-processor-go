package service

import (
	"github.com/gmr458/receipt-processor/domain"
	"github.com/gmr458/receipt-processor/sqlite"
)

type Service struct {
	Receipt domain.ReceiptService
}

func New(sqliteConn *sqlite.Conn) Service {
	return Service{
		Receipt: ReceiptService{sqliteConn},
	}
}
