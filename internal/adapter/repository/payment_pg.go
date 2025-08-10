package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"orderflow/internal/domain/payment"
)

type PaymentPG struct {
	pool *pgxpool.Pool
}

func NewPaymentPG(pool *pgxpool.Pool) *PaymentPG {
	return &PaymentPG{pool: pool}
}

func (r *PaymentPG) CreatePayment(ctx context.Context, paymentEntity *payment.Payment) error {
	const q = `
		INSERT INTO payments (id, order_id, customer_id, amount, currency, status, payment_method, 
		                     transaction_id, failure_reason, processed_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`
	_, err := r.pool.Exec(ctx, q,
		paymentEntity.ID, paymentEntity.OrderID, paymentEntity.CustomerID,
		paymentEntity.Amount, paymentEntity.Currency, string(paymentEntity.Status),
		paymentEntity.PaymentMethod, paymentEntity.TransactionID, paymentEntity.FailureReason,
		paymentEntity.ProcessedAt, paymentEntity.CreatedAt, paymentEntity.UpdatedAt,
	)
	return err
}

func (r *PaymentPG) GetPayment(ctx context.Context, paymentID string) (*payment.Payment, error) {
	const q = `
		SELECT id, order_id, customer_id, amount, currency, status, payment_method,
		       transaction_id, failure_reason, processed_at, created_at, updated_at
		FROM payments WHERE id = $1
	`
	row := r.pool.QueryRow(ctx, q, paymentID)

	var paymentEntity payment.Payment
	var status string
	err := row.Scan(
		&paymentEntity.ID, &paymentEntity.OrderID, &paymentEntity.CustomerID,
		&paymentEntity.Amount, &paymentEntity.Currency, &status,
		&paymentEntity.PaymentMethod, &paymentEntity.TransactionID, &paymentEntity.FailureReason,
		&paymentEntity.ProcessedAt, &paymentEntity.CreatedAt, &paymentEntity.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, payment.NewNotFoundError(paymentID)
	}
	if err != nil {
		return nil, err
	}

	paymentEntity.Status = payment.Status(status)
	return &paymentEntity, nil
}

func (r *PaymentPG) GetPaymentByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	const q = `
		SELECT id, order_id, customer_id, amount, currency, status, payment_method,
		       transaction_id, failure_reason, processed_at, created_at, updated_at
		FROM payments WHERE order_id = $1
	`
	row := r.pool.QueryRow(ctx, q, orderID)

	var paymentEntity payment.Payment
	var status string
	err := row.Scan(
		&paymentEntity.ID, &paymentEntity.OrderID, &paymentEntity.CustomerID,
		&paymentEntity.Amount, &paymentEntity.Currency, &status,
		&paymentEntity.PaymentMethod, &paymentEntity.TransactionID, &paymentEntity.FailureReason,
		&paymentEntity.ProcessedAt, &paymentEntity.CreatedAt, &paymentEntity.UpdatedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, payment.NewNotFoundError("for order " + orderID)
	}
	if err != nil {
		return nil, err
	}

	paymentEntity.Status = payment.Status(status)
	return &paymentEntity, nil
}

func (r *PaymentPG) UpdatePayment(ctx context.Context, paymentEntity *payment.Payment) error {
	const q = `
		UPDATE payments
		SET order_id = $2, customer_id = $3, amount = $4, currency = $5, status = $6,
		    payment_method = $7, transaction_id = $8, failure_reason = $9, processed_at = $10, updated_at = $11
		WHERE id = $1
	`
	ct, err := r.pool.Exec(ctx, q,
		paymentEntity.ID, paymentEntity.OrderID, paymentEntity.CustomerID,
		paymentEntity.Amount, paymentEntity.Currency, string(paymentEntity.Status),
		paymentEntity.PaymentMethod, paymentEntity.TransactionID, paymentEntity.FailureReason,
		paymentEntity.ProcessedAt, paymentEntity.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if ct.RowsAffected() == 0 {
		return payment.NewNotFoundError(paymentEntity.ID)
	}
	return nil
}

func (r *PaymentPG) GetPayments(ctx context.Context) ([]*payment.Payment, error) {
	const q = `
		SELECT id, order_id, customer_id, amount, currency, status, payment_method,
		       transaction_id, failure_reason, processed_at, created_at, updated_at
		FROM payments ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*payment.Payment
	for rows.Next() {
		var paymentEntity payment.Payment
		var status string
		err := rows.Scan(
			&paymentEntity.ID, &paymentEntity.OrderID, &paymentEntity.CustomerID,
			&paymentEntity.Amount, &paymentEntity.Currency, &status,
			&paymentEntity.PaymentMethod, &paymentEntity.TransactionID, &paymentEntity.FailureReason,
			&paymentEntity.ProcessedAt, &paymentEntity.CreatedAt, &paymentEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		paymentEntity.Status = payment.Status(status)
		payments = append(payments, &paymentEntity)
	}

	return payments, rows.Err()
}

func (r *PaymentPG) GetPaymentsByCustomerID(ctx context.Context, customerID string) ([]*payment.Payment, error) {
	const q = `
		SELECT id, order_id, customer_id, amount, currency, status, payment_method,
		       transaction_id, failure_reason, processed_at, created_at, updated_at
		FROM payments WHERE customer_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.pool.Query(ctx, q, customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var payments []*payment.Payment
	for rows.Next() {
		var paymentEntity payment.Payment
		var status string
		err := rows.Scan(
			&paymentEntity.ID, &paymentEntity.OrderID, &paymentEntity.CustomerID,
			&paymentEntity.Amount, &paymentEntity.Currency, &status,
			&paymentEntity.PaymentMethod, &paymentEntity.TransactionID, &paymentEntity.FailureReason,
			&paymentEntity.ProcessedAt, &paymentEntity.CreatedAt, &paymentEntity.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		paymentEntity.Status = payment.Status(status)
		payments = append(payments, &paymentEntity)
	}

	return payments, rows.Err()
}
