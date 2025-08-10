// internal/domain/inventory/errors.go
package inventory

import "fmt"

// InsufficientStockError ошибка недостаточного количества товара
type InsufficientStockError struct {
	ProductID         string
	RequestedQuantity int
	AvailableQuantity int
}

func (e *InsufficientStockError) Error() string {
	return fmt.Sprintf("insufficient stock for product %s: requested %d, available %d", 
		e.ProductID, e.RequestedQuantity, e.AvailableQuantity)
}

func NewInsufficientStockError(productID string, requested, available int) *InsufficientStockError {
	return &InsufficientStockError{
		ProductID:         productID,
		RequestedQuantity: requested,
		AvailableQuantity: available,
	}
}

// ProductNotFoundError ошибка, когда товар не найден
type ProductNotFoundError struct {
	ProductID string
}

func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product not found: %s", e.ProductID)
}

func NewProductNotFoundError(productID string) *ProductNotFoundError {
	return &ProductNotFoundError{ProductID: productID}
}

// ReservationNotFoundError ошибка, когда резервирование не найдено
type ReservationNotFoundError struct {
	OrderID string
}

func (e *ReservationNotFoundError) Error() string {
	return fmt.Sprintf("reservation not found for order: %s", e.OrderID)
}

func NewReservationNotFoundError(orderID string) *ReservationNotFoundError {
	return &ReservationNotFoundError{OrderID: orderID}
}

// ReservationExpiredError ошибка истекшего резервирования
type ReservationExpiredError struct {
	ReservationID string
}

func (e *ReservationExpiredError) Error() string {
	return fmt.Sprintf("reservation expired: %s", e.ReservationID)
}

func NewReservationExpiredError(reservationID string) *ReservationExpiredError {
	return &ReservationExpiredError{ReservationID: reservationID}
}