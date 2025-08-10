package payment

import (
	"context"
)

type Repository interface {
	CreatePayment(ctx context.Context, payment *Payment) error
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)
	UpdatePayment(ctx context.Context, payment *Payment) error
	GetPayments(ctx context.Context) ([]*Payment, error)
	GetPaymentsByCustomerID(ctx context.Context, customerID string) ([]*Payment, error)
}
