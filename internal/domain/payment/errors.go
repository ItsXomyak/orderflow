// internal/domain/payment/errors.go
package payment

import "fmt"

// ValidationError ошибка валидации платежа
type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("payment validation error: %s", e.Message)
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

// NotFoundError ошибка, когда платеж не найден
type NotFoundError struct {
	PaymentID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("payment not found: %s", e.PaymentID)
}

func NewNotFoundError(paymentID string) *NotFoundError {
	return &NotFoundError{PaymentID: paymentID}
}

// ProcessingError ошибка обработки платежа
type ProcessingError struct {
	Code    string
	Message string
}

func (e *ProcessingError) Error() string {
	return fmt.Sprintf("payment processing error [%s]: %s", e.Code, e.Message)
}

func NewProcessingError(code, message string) *ProcessingError {
	return &ProcessingError{Code: code, Message: message}
}

// InsufficientFundsError ошибка недостаточных средств
type InsufficientFundsError struct {
	Amount float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds for amount: %.2f", e.Amount)
}

func NewInsufficientFundsError(amount float64) *InsufficientFundsError {
	return &InsufficientFundsError{Amount: amount}
}

// CannotRefundError ошибка, когда платеж нельзя вернуть
type CannotRefundError struct {
	PaymentID string
	Status    Status
}

func (e *CannotRefundError) Error() string {
	return fmt.Sprintf("cannot refund payment %s with status: %s", e.PaymentID, e.Status)
}

func NewCannotRefundError(paymentID string, status Status) *CannotRefundError {
	return &CannotRefundError{PaymentID: paymentID, Status: status}
}