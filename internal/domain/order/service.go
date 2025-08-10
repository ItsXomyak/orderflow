package order

import "context"

type Service interface {
	Create(ctx context.Context, req *CreateRequest) (*Order, error)

	GetByID(ctx context.Context, id string) (*Order, error)

	Cancel(ctx context.Context, id string) error

	UpdateStatus(ctx context.Context, id string, status Status) error

	SetFailure(ctx context.Context, id string, reason string) error

	Complete(ctx context.Context, id string, paymentID string) error
}
