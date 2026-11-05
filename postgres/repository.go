package postgres

import (
	"github.com/gmr458/receipt-processor/auth"
	"github.com/gmr458/receipt-processor/receipt"
)

type Repository struct {
	Receipt      receipt.ReceiptRepository
	User         auth.UserRepository
	Session      auth.SessionRepository
	Account      auth.AccountRepository
	Verification auth.VerificationRepository
}

func NewRepository(conn *Conn) Repository {
	return Repository{
		Receipt:      ReceiptRepository{conn},
		User:         UserRepository{conn},
		Session:      SessionRepository{conn},
		Account:      AccountRepository{conn},
		Verification: VerificationRepository{conn},
	}
}
