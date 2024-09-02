package repository

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	"github.com/gmr458/receipt-processor/domain"
	"github.com/gmr458/receipt-processor/sqlite"
)

type ReceiptRepository struct {
	sqliteConn *sqlite.Conn
}

func (r ReceiptRepository) FindById(ctx context.Context, id string) (*domain.Receipt, error) {
	tx, err := r.sqliteConn.DB.BeginTx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	queryReceipt := `
        SELECT
            id,
            retailer,
            purchase_date,
            purchase_time,
            total
        FROM receipt
        WHERE id = ?
    `
	receipt := domain.Receipt{Items: []domain.Item{}}
	var timeStr string
	var dateStr string
	row := tx.QueryRow(queryReceipt, id)
	err = row.Scan(
		&receipt.ID,
		&receipt.Retailer,
		&dateStr,
		&timeStr,
		&receipt.Total,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, &domain.Error{Code: domain.ENOTFOUND, Message: "Receipt not found"}
		default:
			return nil, err
		}
	}
	dateParsed, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return nil, err
	}
	timeParsed, err := time.Parse("15:04", timeStr)
	if err != nil {
		return nil, err
	}
	receipt.PurchaseDate = dateParsed
	receipt.PurchaseTime = timeParsed

	queryItems := `
        SELECT
            id,
            short_description,
            price,
            receipt_id
        FROM item
        WHERE receipt_id = ?
    `
	rows, err := tx.QueryContext(ctx, queryItems, receipt.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item domain.Item
		err = rows.Scan(
			&item.ID,
			&item.ShortDescription,
			&item.Price,
			&item.ReceiptID,
		)
		if err != nil {
			return nil, err
		}

		receipt.Items = append(receipt.Items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &receipt, nil
}

func (r ReceiptRepository) Create(ctx context.Context, receipt *domain.Receipt) error {
	tx, err := r.sqliteConn.DB.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback() }()

	queryReceipt := `
        INSERT INTO receipt (
            id,
            retailer,
            purchase_date,
            purchase_time,
            total
        ) VALUES (?, ?, ?, ?, ?)
    `
	args := []any{
		receipt.ID,
		receipt.Retailer,
		receipt.PurchaseDate.Format("2006-01-02"),
		receipt.PurchaseTime.Format("15:04"),
		receipt.Total,
	}
	_, err = tx.ExecContext(ctx, queryReceipt, args...)
	if err != nil {
		return err
	}

	argsItems := []any{}
	var queryItems strings.Builder
	queryItems.WriteString("INSERT INTO item (id, short_description, price, receipt_id) VALUES ")
	for k, v := range receipt.Items {
		if k > 0 {
			queryItems.WriteString(",")
		}
		queryItems.WriteString("(?,?,?,?)")
		argsItems = append(argsItems, v.ID, v.ShortDescription, v.Price, receipt.ID)
	}
	_, err = tx.ExecContext(ctx, queryItems.String(), argsItems...)
	if err != nil {
		return err
	}

	err = tx.Commit()
	if err != nil {
		return err
	}

	return nil
}
