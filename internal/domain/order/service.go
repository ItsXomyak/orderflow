// internal/domain/order/service.go
package order

import "context"

// Service интерфейс для бизнес-логики заказов
type Service interface {
	// Create создает новый заказ
	Create(ctx context.Context, req *CreateRequest) (*Order, error)
	
	// GetByID получает заказ по ID
	GetByID(ctx context.Context, id string) (*Order, error)
	
	// Cancel отменяет заказ
	Cancel(ctx context.Context, id string) error
	
	// UpdateStatus обновляет статус заказа
	UpdateStatus(ctx context.Context, id string, status Status) error
	
	// SetFailure помечает заказ как неудачный
	SetFailure(ctx context.Context, id string, reason string) error
	
	// Complete завершает заказ
	Complete(ctx context.Context, id string, paymentID string) error
}