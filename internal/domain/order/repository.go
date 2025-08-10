package order

import "context"

// Repository интерфейс для работы с заказами
type Repository interface {
	// Create создает новый заказ
	Create(ctx context.Context, order *Order) error
	
	// GetByID получает заказ по ID
	GetByID(ctx context.Context, id string) (*Order, error)
	
	// Update обновляет заказ
	Update(ctx context.Context, order *Order) error
	
	// UpdateStatus обновляет только статус заказа
	UpdateStatus(ctx context.Context, id string, status Status) error
	
	// SetFailure устанавливает статус провала с причиной
	SetFailure(ctx context.Context, id string, reason string) error
	
	// GetByCustomerID получает заказы клиента
	GetByCustomerID(ctx context.Context, customerID string) ([]*Order, error)
	
	// List получает список заказов с пагинацией
	List(ctx context.Context, offset, limit int) ([]*Order, error)
}