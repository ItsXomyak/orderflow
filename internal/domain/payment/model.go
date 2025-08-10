// internal/domain/payment/entity.go
package payment

import "time"

// Status статус платежа
type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRefunded  Status = "refunded"
)

// Payment представляет платеж
type Payment struct {
	ID                string    `json:"id"`
	OrderID           string    `json:"order_id"`
	CustomerID        string    `json:"customer_id"`
	Amount            float64   `json:"amount"`
	Currency          string    `json:"currency"`
	Status            Status    `json:"status"`
	PaymentMethod     string    `json:"payment_method"`
	TransactionID     string    `json:"transaction_id,omitempty"`
	FailureReason     string    `json:"failure_reason,omitempty"`
	ProcessedAt       *time.Time `json:"processed_at,omitempty"`
	CreatedAt         time.Time `json:"created_at"`
	UpdatedAt         time.Time `json:"updated_at"`
}

// Request запрос на обработку платежа
type Request struct {
	OrderID       string  `json:"order_id"`
	CustomerID    string  `json:"customer_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
}

// Response ответ на обработку платежа
type Response struct {
	Success       bool   `json:"success"`
	PaymentID     string `json:"payment_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

// RefundRequest запрос на возврат средств
type RefundRequest struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount,omitempty"` // если не указано, возвращается полная сумма
	Reason    string  `json:"reason"`
}

// NewPayment создает новый платеж
func NewPayment(req *Request) *Payment {
	return &Payment{
		OrderID:       req.OrderID,
		CustomerID:    req.CustomerID,
		Amount:        req.Amount,
		Currency:      req.Currency,
		PaymentMethod: req.PaymentMethod,
		Status:        StatusPending,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// IsCompleted проверяет, завершен ли платеж
func (p *Payment) IsCompleted() bool {
	return p.Status == StatusCompleted
}

// IsFailed проверяет, провален ли платеж
func (p *Payment) IsFailed() bool {
	return p.Status == StatusFailed
}

// IsRefunded проверяет, возвращен ли платеж
func (p *Payment) IsRefunded() bool {
	return p.Status == StatusRefunded
}

// CanBeRefunded проверяет, можно ли вернуть платеж
func (p *Payment) CanBeRefunded() bool {
	return p.Status == StatusCompleted
}

// Complete завершает платеж успешно
func (p *Payment) Complete(transactionID string) {
	p.Status = StatusCompleted
	p.TransactionID = transactionID
	now := time.Now()
	p.ProcessedAt = &now
	p.UpdatedAt = now
}

// Fail помечает платеж как неудачный
func (p *Payment) Fail(reason string) {
	p.Status = StatusFailed
	p.FailureReason = reason
	now := time.Now()
	p.ProcessedAt = &now
	p.UpdatedAt = now
}

// Refund возвращает платеж
func (p *Payment) Refund() error {
	if !p.CanBeRefunded() {
		return NewCannotRefundError(p.ID, p.Status)
	}
	p.Status = StatusRefunded
	p.UpdatedAt = time.Now()
	return nil
}

// Validate валидирует платеж
func (p *Payment) Validate() error {
	if p.OrderID == "" {
		return NewValidationError("order_id is required")
	}
	if p.CustomerID == "" {
		return NewValidationError("customer_id is required")
	}
	if p.Amount <= 0 {
		return NewValidationError("amount must be positive")
	}
	if p.Currency == "" {
		return NewValidationError("currency is required")
	}
	return nil
}