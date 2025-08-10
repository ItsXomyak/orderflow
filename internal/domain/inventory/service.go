// internal/domain/inventory/service.go
package inventory

import "context"

// Service интерфейс для работы со складом
type Service interface {
	// CheckAvailability проверяет наличие товаров
	CheckAvailability(ctx context.Context, req *CheckRequest) (*CheckResponse, error)
	
	// ReserveItems резервирует товары для заказа
	ReserveItems(ctx context.Context, req *ReserveRequest) error
	
	// ReleaseReservation освобождает резервирование товаров
	ReleaseReservation(ctx context.Context, orderID string) error
	
	// ConfirmReservation подтверждает резервирование (переводит в продажу)
	ConfirmReservation(ctx context.Context, orderID string) error
	
	// GetProduct получает информацию о товаре
	GetProduct(ctx context.Context, productID string) (*Product, error)
	
	// UpdateStock обновляет количество товара на складе
	UpdateStock(ctx context.Context, productID string, quantity int) error
}