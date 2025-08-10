package order

import "fmt"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("order validation error: %s", e.Message)
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

type NotFoundError struct {
	OrderID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("order not found: %s", e.OrderID)
}

func NewNotFoundError(orderID string) *NotFoundError {
	return &NotFoundError{OrderID: orderID}
}

type CannotCancelError struct {
	Status Status
}

func (e *CannotCancelError) Error() string {
	return fmt.Sprintf("cannot cancel order with status: %s", e.Status)
}

func NewCannotCancelError(status Status) *CannotCancelError {
	return &CannotCancelError{Status: status}
}

type StatusTransitionError struct {
	FromStatus Status
	ToStatus   Status
}

func (e *StatusTransitionError) Error() string {
	return fmt.Sprintf("invalid status transition from %s to %s", e.FromStatus, e.ToStatus)
}

func NewStatusTransitionError(from, to Status) *StatusTransitionError {
	return &StatusTransitionError{FromStatus: from, ToStatus: to}
}
