package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/gmr458/receipt-processor/domain"
)

type ReceiptService struct {
	repository domain.ReceiptRepository
	cache      domain.ReceiptCache
}

func NewReceiptService(repository domain.ReceiptRepository, cache domain.ReceiptCache) ReceiptService {
	return ReceiptService{
		repository,
		cache,
	}
}

func (s *ReceiptService) Process(ctx context.Context, dto domain.ReceiptDTO) (*domain.Receipt, error) {
	isValid, errors := dto.IsValid()
	if !isValid {
		return nil, &domain.Error{
			Code:    domain.EINVALID,
			Message: "Invalid field/s",
			Details: errors,
		}
	}

	receipt := &domain.Receipt{
		ID:       uuid.New().String(),
		Retailer: dto.Retailer,
		Total:    dto.Total,
		Items:    []domain.Item{},
	}
	parsedDate, err := time.Parse("2006-01-02", dto.PurchaseDate)
	if err != nil {
		return nil, &domain.Error{
			Code:    domain.EINVALID,
			Message: "Invalid field/s",
			Details: map[string]string{
				"purchaseDate": "invalid format, it should be YYYY-MM-DD",
			},
		}
	}
	receipt.PurchaseDate = parsedDate

	parsedTime, err := time.Parse("15:04", dto.PurchaseTime)
	if err != nil {
		return nil, &domain.Error{
			Code:    domain.EINVALID,
			Message: "Invalid field/s",
			Details: map[string]string{
				"purchaseTime": "invalid format, it should be hh:mm",
			},
		}
	}
	receipt.PurchaseTime = parsedTime

	for _, itemDto := range dto.Items {
		item := domain.Item{
			ID:               uuid.New().String(),
			ShortDescription: itemDto.ShortDescription,
			Price:            itemDto.Price,
		}
		receipt.Items = append(receipt.Items, item)
	}

	err = s.repository.Create(ctx, receipt)
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

func (s *ReceiptService) GetReceipts(
	ctx context.Context,
	filters domain.Filters,
) (domain.PaginatedReceipts, error) {
	isValid, errors := filters.IsValid()
	if !isValid {
		return domain.PaginatedReceipts{}, &domain.Error{
			Code:    domain.EINVALID,
			Message: "Invalid filter params",
			Details: errors,
		}
	}

	key := fmt.Sprintf(
		"receipts:page:%d:limit:%d:sort:%s",
		filters.Page,
		filters.Limit,
		filters.Sort,
	)

	paginatedReceipts, err := s.cache.GetPaginatedReceipts(ctx, key)
	if nil == err {
		return paginatedReceipts, nil
	}

	paginatedReceipts, err = s.repository.Find(ctx, filters)
	if err != nil {
		return domain.PaginatedReceipts{}, err
	}

	_ = s.cache.SetPaginatedReceipts(ctx, key, paginatedReceipts)

	return paginatedReceipts, nil
}
