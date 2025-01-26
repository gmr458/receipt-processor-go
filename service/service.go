package service

import (
	"github.com/gmr458/receipt-processor/cache"
	"github.com/gmr458/receipt-processor/repository"
)

type Service struct {
	Receipt ReceiptService
}

func New(repository repository.Repository, cache cache.Cache) Service {
	return Service{
		Receipt: NewReceiptService(
			repository.Receipt,
			cache.Receipt,
		),
	}
}
