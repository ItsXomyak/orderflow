package inventory

import "fmt"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("inventory validation error: %s", e.Message)
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

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

type ProductNotFoundError struct {
	ProductID string
}

func (e *ProductNotFoundError) Error() string {
	return fmt.Sprintf("product not found: %s", e.ProductID)
}

func NewProductNotFoundError(productID string) *ProductNotFoundError {
	return &ProductNotFoundError{ProductID: productID}
}

type ReservationNotFoundError struct {
	OrderID string
}

func (e *ReservationNotFoundError) Error() string {
	return fmt.Sprintf("reservation not found for order: %s", e.OrderID)
}

func NewReservationNotFoundError(orderID string) *ReservationNotFoundError {
	return &ReservationNotFoundError{OrderID: orderID}
}

type ReservationExpiredError struct {
	ReservationID string
}

func (e *ReservationExpiredError) Error() string {
	return fmt.Sprintf("reservation expired: %s", e.ReservationID)
}

func NewReservationExpiredError(reservationID string) *ReservationExpiredError {
	return &ReservationExpiredError{ReservationID: reservationID}
}
