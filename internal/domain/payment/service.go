package payment

import (
	"context"
	"time"
)

type Service interface {
	ProcessPayment(ctx context.Context, req *Request) (*Response, error)

	GetPayment(ctx context.Context, paymentID string) (*Payment, error)

	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)

	RefundPayment(ctx context.Context, req *RefundRequest) error

	CancelPayment(ctx context.Context, paymentID string) error
}

type Gateway interface {
	Charge(ctx context.Context, req *Request) (*Response, error)

	Refund(ctx context.Context, transactionID string, amount float64) error

	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
}

type Transaction struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
