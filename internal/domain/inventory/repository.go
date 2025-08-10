package inventory

import (
	"context"
)

type ReservationID string

type Repository interface {
	CreateProduct(ctx context.Context, product *Product) error
	GetProduct(ctx context.Context, productID string) (*Product, error)
	UpdateProduct(ctx context.Context, product *Product) error
	GetProducts(ctx context.Context) ([]*Product, error)
	
	CreateReservation(ctx context.Context, reservation *Reservation) error
	GetReservationByOrderID(ctx context.Context, orderID string) (*Reservation, error)
	DeleteReservation(ctx context.Context, orderID string) error
	GetExpiredReservations(ctx context.Context) ([]*Reservation, error)
}
