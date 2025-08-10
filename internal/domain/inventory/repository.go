package inventory

import (
	"context"

	"orderflow/internal/domain/order"
)

type ReservationID string

type Repository interface {
	CheckAndReserve(ctx context.Context, items []order.Item, orderID order.Order) (ReservationID, error) // проверить это место
	Release(ctx context.Context, reservationID ReservationID) error
	Commit(ctx context.Context, reservationID ReservationID) error
}
