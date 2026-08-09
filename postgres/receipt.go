package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/gmr458/receipt-processor/errs"
	"github.com/gmr458/receipt-processor/receipt"
)

type ReceiptRepository struct {
	conn *Conn
}

func (r ReceiptRepository) FindById(ctx context.Context, id string) (*receipt.Receipt, error) {
	tx, err := r.conn.Pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queryReceipt := `
        SELECT
            id,
            retailer,
            purchase_date,
            purchase_time,
            total
        FROM receipt
        WHERE id = $1
    `
	rec := receipt.Receipt{Items: []receipt.Item{}}
	row := tx.QueryRow(ctx, queryReceipt, id)
	err = row.Scan(
		&rec.ID,
		&rec.Retailer,
		&rec.PurchaseDate,
		&rec.PurchaseTime,
		&rec.Total,
	)
	if err != nil {
		switch {
		case errors.Is(err, pgx.ErrNoRows):
			return nil, &errs.Error{Code: errs.ENOTFOUND, Message: "Receipt not found"}
		default:
			return nil, err
		}
	}

	queryItems := `
        SELECT
            id,
            short_description,
            price,
            receipt_id
        FROM item
        WHERE receipt_id = $1
    `
	rows, err := tx.Query(ctx, queryItems, rec.ID)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var item receipt.Item
		err = rows.Scan(
			&item.ID,
			&item.ShortDescription,
			&item.Price,
			&item.ReceiptID,
		)
		if err != nil {
			rows.Close()
			return nil, err
		}

		rec.Items = append(rec.Items, item)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return nil, err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return nil, err
	}

	return &rec, nil
}

func (r ReceiptRepository) Create(ctx context.Context, rec *receipt.Receipt) error {
	tx, err := r.conn.Pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	queryReceipt := `
        INSERT INTO receipt (
            id,
            retailer,
            purchase_date,
            purchase_time,
            total
        ) VALUES ($1, $2, $3, $4, $5)
    `
	_, err = tx.Exec(ctx, queryReceipt,
		rec.ID,
		rec.Retailer,
		rec.PurchaseDate,
		rec.PurchaseTime,
		rec.Total,
	)
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
			offset := k * 4
			queryItems.WriteString(fmt.Sprintf("($%d,$%d,$%d,$%d)", offset+1, offset+2, offset+3, offset+4))
			argsItems = append(argsItems, v.ID, v.ShortDescription, v.Price, rec.ID)
		}
		_, err = tx.Exec(ctx, queryItems.String(), argsItems...)
		if err != nil {
			return err
		}
	}

	err = tx.Commit(ctx)
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
	err := r.conn.Pool.QueryRow(ctx, "SELECT count(*) FROM receipt").Scan(&total)
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
        LIMIT $1 OFFSET $2`,
		filters.SortColumn(),
		filters.SortDirection(),
	)

	rows, err := r.conn.Pool.Query(
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
		err = rows.Scan(
			&rec.ID,
			&rec.Retailer,
			&rec.PurchaseDate,
			&rec.PurchaseTime,
			&rec.Total,
		)
		if err != nil {
			rows.Close()
			return receipt.PaginatedReceipts{}, err
		}
		rec.Items = []receipt.Item{}
		receipts = append(receipts, rec)
		receiptIDs = append(receiptIDs, rec.ID)
	}
	err = rows.Err()
	rows.Close()
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
		pgPlaceholders(1, len(receiptIDs)),
	)

	args := make([]any, len(receiptIDs))
	for i, id := range receiptIDs {
		args[i] = id
	}

	itemRows, err := r.conn.Pool.Query(ctx, queryItems, args...)
	if err != nil {
		return receipt.PaginatedReceipts{}, err
	}

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
			itemRows.Close()
			return receipt.PaginatedReceipts{}, err
		}
		itemsByReceiptID[item.ReceiptID] = append(itemsByReceiptID[item.ReceiptID], item)
	}
	err = itemRows.Err()
	itemRows.Close()
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

// pgPlaceholders produces "$1,$2,..." starting at the given offset.
func pgPlaceholders(start, count int) string {
	placeholders := make([]string, 0, count)
	for i := 0; i < count; i++ {
		placeholders = append(placeholders, fmt.Sprintf("$%d", start+i))
	}
	return strings.Join(placeholders, ",")
}
