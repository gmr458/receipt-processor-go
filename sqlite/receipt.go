package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gmr458/receipt-processor-go/errs"
	"github.com/gmr458/receipt-processor-go/money"
	"github.com/gmr458/receipt-processor-go/receipt"
	"github.com/gmr458/receipt-processor-go/timefmt"
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
	var totalCents int64
	row := tx.QueryRow(queryReceipt, id)
	err = row.Scan(
		&rec.ID,
		&rec.Retailer,
		&dateStr,
		&timeStr,
		&totalCents,
	)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "Receipt not found"}
		default:
			return nil, err
		}
	}
	dateParsed, err := time.Parse(time.DateOnly, dateStr)
	if err != nil {
		return nil, err
	}
	timeParsed, err := time.Parse(timefmt.TimeHourMinuteOnly, timeStr)
	if err != nil {
		return nil, err
	}
	rec.PurchaseDate = dateParsed
	rec.PurchaseTime = timeParsed
	rec.Total = money.FromCents(totalCents)

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

	for rows.Next() {
		var item receipt.Item
		var priceCents int64
		err = rows.Scan(
			&item.ID,
			&item.ShortDescription,
			&priceCents,
			&item.ReceiptID,
		)
		if err != nil {
			_ = rows.Close()
			return nil, err
		}
		item.Price = money.FromCents(priceCents)

		rec.Items = append(rec.Items, item)
	}

	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r ReceiptRepository) Create(ctx context.Context, rec *receipt.Receipt) error {
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
		rec.ID,
		rec.Retailer,
		rec.PurchaseDate.Format(time.DateOnly),
		rec.PurchaseTime.Format(timefmt.TimeHourMinuteOnly),
		int64(rec.Total),
	}
	_, err = tx.ExecContext(ctx, queryReceipt, args...)
	if err != nil {
		return err
	}

	if len(rec.Items) > 0 {
		argsItems := make([]any, 0, len(rec.Items)*4)
		var queryItems strings.Builder
		queryItems.Grow(70 + (len(rec.Items) * 10))
		queryItems.WriteString("INSERT INTO item (id, short_description, price, receipt_id) VALUES ")
		for k, v := range rec.Items {
			if k > 0 {
				queryItems.WriteString(",")
			}
			queryItems.WriteString("(?,?,?,?)")
			argsItems = append(argsItems, v.ID, v.ShortDescription, int64(v.Price), rec.ID)
		}
		_, err = tx.ExecContext(ctx, queryItems.String(), argsItems...)
		if err != nil {
			return err
		}
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
	tx, err := r.conn.DB.BeginTx(ctx, nil)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}
	defer func() { _ = tx.Rollback() }()

	var total int
	err = tx.QueryRowContext(ctx, "SELECT count(*) FROM receipt").Scan(&total)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	sortColumn, err := filters.SortColumn()
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
		sortColumn,
		filters.SortDirection(),
	)

	rows, err := tx.QueryContext(
		ctx,
		queryReceipts,
		filters.Limit,
		filters.Offset(),
	)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	receipts := make([]receipt.Receipt, 0, filters.Limit)
	receiptIDs := make([]string, 0, filters.Limit)

	for rows.Next() {
		var rec receipt.Receipt
		var timeStr string
		var dateStr string
		var totalCents int64
		err = rows.Scan(
			&rec.ID,
			&rec.Retailer,
			&dateStr,
			&timeStr,
			&totalCents,
		)
		if err != nil {
			_ = rows.Close()
			return receipt.PaginatedReceipts{}, err
		}
		dateParsed, err := time.Parse(time.DateOnly, dateStr)
		if err != nil {
			_ = rows.Close()
			return receipt.PaginatedReceipts{}, err
		}
		timeParsed, err := time.Parse(timefmt.TimeHourMinuteOnly, timeStr)
		if err != nil {
			_ = rows.Close()
			return receipt.PaginatedReceipts{}, err
		}
		rec.PurchaseDate = dateParsed
		rec.PurchaseTime = timeParsed
		rec.Total = money.FromCents(totalCents)
		rec.Items = []receipt.Item{}
		receipts = append(receipts, rec)
		receiptIDs = append(receiptIDs, rec.ID)
	}

	err = rows.Err()
	_ = rows.Close()
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	itemsByReceiptID := make(map[string][]receipt.Item, filters.Limit)
	if len(receiptIDs) > 0 {
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

		itemRows, err := tx.QueryContext(ctx, queryItems, args...)
		if err != nil {
			return receipt.PaginatedReceipts{}, err
		}

		for itemRows.Next() {
			var item receipt.Item
			var priceCents int64
			err = itemRows.Scan(
				&item.ID,
				&item.ShortDescription,
				&priceCents,
				&item.ReceiptID,
			)
			if err != nil {
				_ = itemRows.Close()
				return receipt.PaginatedReceipts{}, err
			}
			item.Price = money.FromCents(priceCents)
			itemsByReceiptID[item.ReceiptID] = append(itemsByReceiptID[item.ReceiptID], item)
		}

		err = itemRows.Err()
		_ = itemRows.Close()
		if err != nil {
			return receipt.PaginatedReceipts{}, err
		}
	}

	for i := range receipts {
		if items, ok := itemsByReceiptID[receipts[i].ID]; ok {
			receipts[i].Items = items
		}
	}

	err = tx.Commit()
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

	metadata := receipt.CalculateMetadata(total, filters.Page, filters.Limit)

	return receipt.PaginatedReceipts{
		Receipts: receipts,
		Metadata: &metadata,
	}, nil
}
