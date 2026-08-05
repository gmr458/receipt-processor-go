package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gmr458/receipt-processor/errs"
	"github.com/gmr458/receipt-processor/receipt"
)

type ReceiptRepository struct {
	conn *Conn
}

func (r ReceiptRepository) FindById(ctx context.Context, id string) (*receipt.Receipt, error) {
	tx, err := r.conn.DB.BeginTx(ctx, nil)
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
	rec := receipt.Receipt{Items: []receipt.Item{}}
	var timeStr string
	var dateStr string
	row := tx.QueryRow(queryReceipt, id)
	err = row.Scan(
		&rec.ID,
		&rec.Retailer,
		&dateStr,
		&timeStr,
		&rec.Total,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "Receipt not found"}
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
	rec.PurchaseDate = dateParsed
	rec.PurchaseTime = timeParsed

	queryItems := `
        SELECT
            id,
            short_description,
            price,
            receipt_id
        FROM item
        WHERE receipt_id = ?
    `
	rows, err := tx.QueryContext(ctx, queryItems, rec.ID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item receipt.Item
		err = rows.Scan(
			&item.ID,
			&item.ShortDescription,
			&item.Price,
			&item.ReceiptID,
		)
		if err != nil {
			return nil, err
		}

		rec.Items = append(rec.Items, item)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r ReceiptRepository) Create(ctx context.Context, receipt *receipt.Receipt) error {
	tx, err := r.conn.DB.BeginTx(ctx, nil)
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

	argsItems := make([]any, 0, len(receipt.Items)*4)
	var queryItems strings.Builder
	queryItems.Grow(70 + (len(receipt.Items) * 10))
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

func (r ReceiptRepository) Find(
	ctx context.Context,
	filters receipt.Filters,
) (receipt.PaginatedReceipts, error) {
	var total int
	err := r.conn.DB.QueryRowContext(ctx, "SELECT count(*) FROM receipt").Scan(&total)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	queryReceipts := fmt.Sprintf(
		`SELECT
            id,
            retailer,
            purchase_date,
            purchase_time,
            total
        FROM receipt
        ORDER BY %s %s
        LIMIT ? OFFSET ?`,
		filters.SortColumn(),
		filters.SortDirection(),
	)

	rows, err := r.conn.DB.QueryContext(
		ctx,
		queryReceipts,
		filters.Limit,
		filters.Offset(),
	)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}
	defer rows.Close()

	receipts := make([]receipt.Receipt, 0, filters.Limit)
	receiptIDs := make([]string, 0, filters.Limit)

	for rows.Next() {
		var rec receipt.Receipt
		var timeStr string
		var dateStr string
		err = rows.Scan(
			&rec.ID,
			&rec.Retailer,
			&dateStr,
			&timeStr,
			&rec.Total,
		)
		if err != nil {
			return receipt.PaginatedReceipts{}, err
		}
		dateParsed, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			return receipt.PaginatedReceipts{}, err
		}
		timeParsed, err := time.Parse("15:04", timeStr)
		if err != nil {
			return receipt.PaginatedReceipts{}, err
		}
		rec.PurchaseDate = dateParsed
		rec.PurchaseTime = timeParsed
		rec.Items = []receipt.Item{}
		receipts = append(receipts, rec)
		receiptIDs = append(receiptIDs, rec.ID)
	}

	err = rows.Err()
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	if len(receipts) == 0 {
		return receipt.PaginatedReceipts{
			Receipts: receipts,
			Metadata: nil,
		}, nil
	}

	queryItems := fmt.Sprintf(
		`SELECT
            id,
            short_description,
            price,
            receipt_id
        FROM item
        WHERE receipt_id IN (%s)`,
		strings.Repeat("?,", len(receiptIDs)-1)+"?",
	)

	args := make([]any, len(receiptIDs))
	for i, id := range receiptIDs {
		args[i] = id
	}

	itemRows, err := r.conn.DB.QueryContext(ctx, queryItems, args...)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}
	defer itemRows.Close()

	itemsByReceiptID := make(map[string][]receipt.Item, filters.Limit)
	for itemRows.Next() {
		var item receipt.Item
		err = itemRows.Scan(
			&item.ID,
			&item.ShortDescription,
			&item.Price,
			&item.ReceiptID,
		)
		if err != nil {
			return receipt.PaginatedReceipts{}, err
		}
		itemsByReceiptID[item.ReceiptID] = append(itemsByReceiptID[item.ReceiptID], item)
	}

	err = itemRows.Err()
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	for i := range receipts {
		if items, ok := itemsByReceiptID[receipts[i].ID]; ok {
			receipts[i].Items = items
		}
	}

	metadata := receipt.CalculateMetadata(total, filters.Page, filters.Limit)

	return receipt.PaginatedReceipts{
		Receipts: receipts,
		Metadata: &metadata,
	}, nil
}
