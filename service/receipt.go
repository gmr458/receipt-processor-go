package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/gmr458/receipt-processor/domain"
)

type ReceiptService struct {
	repository domain.ReceiptRepository
	cache      domain.ReceiptCache
}

func NewReceiptService(repository domain.ReceiptRepository, cache domain.ReceiptCache) *ReceiptService {
	return &ReceiptService{
		repository,
		cache,
	}
}

func (s *ReceiptService) Process(ctx context.Context, dto *domain.ReceiptDTO) (*domain.Receipt, error) {
	receipt := &domain.Receipt{
		ID:       uuid.New().String(),
		Retailer: dto.Retailer,
		Total:    dto.Total,
		Items:    []domain.Item{},
	}
	receipt.PurchaseDate, _ = time.Parse("2006-01-02", dto.PurchaseDate)
	receipt.PurchaseTime, _ = time.Parse("15:04", dto.PurchaseTime)

	for _, itemDto := range dto.Items {
		item := domain.Item{
			ID:               uuid.New().String(),
			ShortDescription: itemDto.ShortDescription,
			Price:            itemDto.Price,
		}
		receipt.Items = append(receipt.Items, item)
	}

	err := s.repository.Create(ctx, receipt)
	if err != nil {
		return nil, err
	}

	err = s.cache.SetPointsById(ctx, receipt.ID, receipt.CalculateTotalPoints())
	if err != nil {
		return nil, err
	}

	return receipt, nil
}

func (s *ReceiptService) GetPointsById(ctx context.Context, id string) (int, error) {
	err := uuid.Validate(id)
	if err != nil {
		return 0, &domain.Error{Code: domain.ENOTFOUND, Message: "Receipt not found"}
	}

	points, err := s.cache.GetPointsById(ctx, id)
	if nil == err {
		return points, nil
	}

	receipt, err := s.repository.FindById(ctx, id)
	if err != nil {
		return 0, err
	}

	points = receipt.CalculateTotalPoints()

	err = s.cache.SetPointsById(ctx, receipt.ID, points)
	if err != nil {
		return 0, err
	}

	return points, nil
}
