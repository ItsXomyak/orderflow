package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"orderflow/internal/domain/order"
)

type OrderPG struct {
	pool *pgxpool.Pool
}

func NewOrderPG(pool *pgxpool.Pool) *OrderPG { return &OrderPG{pool: pool} }

// Create: транзакция — вставить orders + order_items
func (r *OrderPG) Create(ctx context.Context, o *order.Order) error {
	tx, err := r.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()

	const qOrder = `
		INSERT INTO orders (id, customer_id, status, total_amount, payment_id, failure_reason, created_at, updated_at, completed_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)
	`
	_, err = tx.Exec(ctx, qOrder,
		o.ID, o.CustomerID, string(o.Status), o.TotalAmount, nil, nil, o.CreatedAt, o.UpdatedAt, o.CompletedAt,
	)
	if err != nil {
		return err
	}

	b := &pgx.Batch{}
	const qItem = `
		INSERT INTO order_items (order_id, product_id, name, quantity, price)
		VALUES ($1,$2,$3,$4,$5)
	`
	for _, it := range o.Items {
		b.Queue(qItem, o.ID, it.ProductID, it.Name, it.Quantity, it.Price)
	}
	br := tx.SendBatch(ctx, b)
	if err := br.Close(); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

func (r *OrderPG) GetByID(ctx context.Context, id string) (*order.Order, error) {
	const qOrder = `
		SELECT id, customer_id, status, total_amount, payment_id, failure_reason, created_at, updated_at, completed_at
		FROM orders WHERE id=$1
	`
	row := r.pool.QueryRow(ctx, qOrder, id)

	var o order.Order
	var status string
	err := row.Scan(&o.ID, &o.CustomerID, &status, &o.TotalAmount, &o.PaymentID, &o.FailureReason, &o.CreatedAt, &o.UpdatedAt, &o.CompletedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, order.NewNotFoundError(id)
	}
	if err != nil {
		return nil, err
	}
	o.Status = order.Status(status)

	const qItems = `
		SELECT product_id, name, quantity, price
		FROM order_items WHERE order_id=$1 ORDER BY id
	`
	rows, err := r.pool.Query(ctx, qItems, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var it order.Item
		if err := rows.Scan(&it.ProductID, &it.Name, &it.Quantity, &it.Price); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, it)
	}
	return &o, rows.Err()
}

func (r *OrderPG) Update(ctx context.Context, o *order.Order) error {
	const q = `
		UPDATE orders
		SET customer_id=$2, status=$3, total_amount=$4, payment_id=$5, failure_reason=$6, updated_at=$7, completed_at=$8
		WHERE id=$1
	`
	_, err := r.pool.Exec(ctx, q,
		o.ID, o.CustomerID, string(o.Status), o.TotalAmount, o.PaymentID, o.FailureReason, time.Now(), o.CompletedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return order.NewNotFoundError(o.ID)
	}
	return err
}

func (r *OrderPG) UpdateStatus(ctx context.Context, id string, st order.Status) error {
	const q = `UPDATE orders SET status=$2, updated_at=NOW() WHERE id=$1`
	ct, err := r.pool.Exec(ctx, q, id, string(st))
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return order.NewNotFoundError(id)
	}
	return nil
}

func (r *OrderPG) SetFailure(ctx context.Context, id, reason string) error {
	const q = `UPDATE orders SET status='failed', failure_reason=$2, updated_at=NOW() WHERE id=$1`
	ct, err := r.pool.Exec(ctx, q, id, reason)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return order.NewNotFoundError(id)
	}
	return nil
}

func (r *OrderPG) GetByCustomerID(ctx context.Context, customerID string) ([]*order.Order, error) {
	const q = `SELECT id FROM orders WHERE customer_id=$1 ORDER BY created_at DESC`
	rows, err := r.pool.Query(ctx, q, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*order.Order
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		o, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		res = append(res, o)
	}
	return res, rows.Err()
}

func (r *OrderPG) List(ctx context.Context, offset, limit int) ([]*order.Order, error) {
	if limit <= 0 {
		limit = 50
	}
	const q = `SELECT id FROM orders ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.pool.Query(ctx, q, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var res []*order.Order
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		o, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		res = append(res, o)
	}
	return res, rows.Err()
}
