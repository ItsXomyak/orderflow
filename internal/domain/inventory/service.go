package inventory

import "context"

type Service interface {
	CheckAvailability(ctx context.Context, req *CheckRequest) (*CheckResponse, error)
	
	ReserveItems(ctx context.Context, req *ReserveRequest) error
	
	ReleaseReservation(ctx context.Context, orderID string) error
	
	ConfirmReservation(ctx context.Context, orderID string) error
	
	GetProduct(ctx context.Context, productID string) (*Product, error)
	
	UpdateStock(ctx context.Context, productID string, quantity int) error
}