package inventory

import (
	"context"

	"orderflow/internal/domain/order"
)

type ReservationID string

type Repository interface {
	CheckAndReserve(ctx context.Context, items []order.Item, orderID order.OrderID) (ReservationID, error)
	Release(ctx context.Context, reservationID ReservationID) error
	Commit(ctx context.Context, reservationID ReservationID) error
}
