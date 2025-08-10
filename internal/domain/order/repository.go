package order

import "context"

type Repository interface {
	Save(ctx context.Context, o *Order) error
	Get(ctx context.Context, id OrderID) (*Order, error)
}
