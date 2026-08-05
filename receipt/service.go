package receipt

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/gmr458/receipt-processor/errs"
)

type Service struct {
	repository ReceiptRepository
	cache      ReceiptCache
}

func NewService(repository ReceiptRepository, cache ReceiptCache) Service {
	return Service{
		repository,
		cache,
	}
}

func (s *Service) Process(ctx context.Context, dto ReceiptDTO) (*Receipt, error) {
	isValid, errors := dto.IsValid()
	if !isValid {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: errors,
		}
	}

	rec := &Receipt{
		ID:       uuid.New().String(),
		Retailer: dto.Retailer,
		Total:    dto.Total,
		Items:    make([]Item, 0, len(dto.Items)),
	}
	parsedDate, err := time.Parse("2006-01-02", dto.PurchaseDate)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: map[string]string{
				"purchaseDate": "invalid format, it should be YYYY-MM-DD",
			},
		}
	}
	rec.PurchaseDate = parsedDate

	parsedTime, err := time.Parse("15:04", dto.PurchaseTime)
	if err != nil {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field/s",
			Details: map[string]string{
				"purchaseTime": "invalid format, it should be hh:mm",
			},
		}
	}
	rec.PurchaseTime = parsedTime

	for _, itemDto := range dto.Items {
		item := Item{
			ID:               uuid.New().String(),
			ShortDescription: itemDto.ShortDescription,
			Price:            itemDto.Price,
		}
		rec.Items = append(rec.Items, item)
	}

	err = s.repository.Create(ctx, rec)
	if err != nil {
		return nil, err
	}

	go func() {
		_ = s.cache.SetPointsById(
			context.Background(),
			rec.ID,
			rec.CalculateTotalPoints(),
			5*time.Minute,
		)
	}()

	return rec, nil
}

func (s *Service) GetPointsById(ctx context.Context, id string) (int, error) {
	err := uuid.Validate(id)
	if err != nil {
		return 0, &errs.Error{Code: errs.ENOTFOUND, Message: "Receipt not found"}
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

	go func() {
		_ = s.cache.SetPointsById(
			context.Background(),
			receipt.ID,
			points,
			5*time.Minute,
		)
	}()

	return points, nil
}

func (s *Service) GetReceipts(
	ctx context.Context,
	filters Filters,
) (PaginatedReceipts, error) {
	isValid, errors := filters.IsValid()
	if !isValid {
		return PaginatedReceipts{}, &errs.Error{
			Code:    errs.EINVALID,
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
		return PaginatedReceipts{}, err
	}

	go func() {
		_ = s.cache.SetPaginatedReceipts(
			context.Background(),
			key,
			paginatedReceipts,
			5*time.Minute,
		)
	}()

	return paginatedReceipts, nil
}
