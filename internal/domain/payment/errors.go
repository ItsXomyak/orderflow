package payment

import "fmt"

type ValidationError struct {
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("payment validation error: %s", e.Message)
}

func NewValidationError(message string) *ValidationError {
	return &ValidationError{Message: message}
}

type NotFoundError struct {
	PaymentID string
}

func (e *NotFoundError) Error() string {
	return fmt.Sprintf("payment not found: %s", e.PaymentID)
}

func NewNotFoundError(paymentID string) *NotFoundError {
	return &NotFoundError{PaymentID: paymentID}
}

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

type InsufficientFundsError struct {
	Amount float64
}

func (e *InsufficientFundsError) Error() string {
	return fmt.Sprintf("insufficient funds for amount: %.2f", e.Amount)
}

func NewInsufficientFundsError(amount float64) *InsufficientFundsError {
	return &InsufficientFundsError{Amount: amount}
}

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

type DuplicatePaymentError struct {
	OrderID string
}

func (e *DuplicatePaymentError) Error() string {
	return fmt.Sprintf("duplicate payment for order: %s", e.OrderID)
}

func NewDuplicatePaymentError(orderID string) *DuplicatePaymentError {
	return &DuplicatePaymentError{OrderID: orderID}
}

type CannotCancelError struct {
	PaymentID string
	Status    Status
}

func (e *CannotCancelError) Error() string {
	return fmt.Sprintf("cannot cancel payment %s with status: %s", e.PaymentID, e.Status)
}

func NewCannotCancelError(paymentID string, status Status) *CannotCancelError {
	return &CannotCancelError{PaymentID: paymentID, Status: status}
}

type RefundFailedError struct {
	PaymentID string
	Reason    string
}

func (e *RefundFailedError) Error() string {
	return fmt.Sprintf("refund failed for payment %s: %s", e.PaymentID, e.Reason)
}

func NewRefundFailedError(paymentID, reason string) *RefundFailedError {
	return &RefundFailedError{PaymentID: paymentID, Reason: reason}
}