package receipt

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/gmr458/receipt-processor-go/errs"
	"github.com/gmr458/receipt-processor-go/timefmt"
)

type Service struct {
	repository ReceiptRepository
}

func NewService(repository ReceiptRepository) Service {
	return Service{
		repository,
	}
}

func (s *Service) Process(ctx context.Context, dto ReceiptDTO) (*Receipt, error) {
	isValid, validationErrors := dto.IsValid()
	if !isValid {
		return nil, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid field(s)",
			Details: validationErrors,
		}
	}

	rec := &Receipt{
		ID:       uuid.New().String(),
		Retailer: dto.Retailer,
		Total:    dto.Total,
		Items:    make([]Item, 0, len(dto.Items)),
	}

	var err error
	rec.PurchaseDate, err = time.Parse(time.DateOnly, dto.PurchaseDate)
	if err != nil {
		return nil, errs.Errorf(errs.EINTERNAL, "validated purchaseDate %q failed to parse: %v", dto.PurchaseDate, err)
	}

	rec.PurchaseTime, err = time.Parse(timefmt.TimeHourMinuteOnly, dto.PurchaseTime)
	if err != nil {
		return nil, errs.Errorf(errs.EINTERNAL, "validated purchaseTime %q failed to parse: %v", dto.PurchaseTime, err)
	}

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

	return rec, nil
}

func (s *Service) GetPointsById(ctx context.Context, id string) (int, error) {
	err := uuid.Validate(id)
	if err != nil {
		return 0, &errs.Error{Code: errs.EINVALID, Message: "Invalid receipt id"}
	}

	receipt, err := s.repository.FindById(ctx, id)
	if err != nil {
		return 0, err
	}

	return receipt.CalculateTotalPoints(), nil
}

func (s *Service) GetReceipts(
	ctx context.Context,
	filters Filters,
) (PaginatedReceipts, error) {
	isValid, validationErrors := filters.IsValid()
	if !isValid {
		return PaginatedReceipts{}, &errs.Error{
			Code:    errs.EINVALID,
			Message: "Invalid filter params",
			Details: validationErrors,
		}
	}

	return s.repository.Find(ctx, filters)
}
