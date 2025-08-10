package order

import "context"

type Repository interface {
	Create(ctx context.Context, order *Order) error

	GetByID(ctx context.Context, id string) (*Order, error)

	Update(ctx context.Context, order *Order) error

	UpdateStatus(ctx context.Context, id string, status Status) error

	SetFailure(ctx context.Context, id string, reason string) error

	GetByCustomerID(ctx context.Context, customerID string) ([]*Order, error)

	List(ctx context.Context, offset, limit int) ([]*Order, error)
}
