package payment

import "time"

type Status string

const (
	StatusPending   Status = "pending"
	StatusCompleted Status = "completed"
	StatusFailed    Status = "failed"
	StatusRefunded  Status = "refunded"
)

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

type Request struct {
	OrderID       string  `json:"order_id"`
	CustomerID    string  `json:"customer_id"`
	Amount        float64 `json:"amount"`
	Currency      string  `json:"currency"`
	PaymentMethod string  `json:"payment_method"`
}

type Response struct {
	Success       bool   `json:"success"`
	PaymentID     string `json:"payment_id,omitempty"`
	TransactionID string `json:"transaction_id,omitempty"`
	ErrorCode     string `json:"error_code,omitempty"`
	ErrorMessage  string `json:"error_message,omitempty"`
}

type RefundRequest struct {
	PaymentID string  `json:"payment_id"`
	Amount    float64 `json:"amount,omitempty"` // если не указано, возвращается полная сумма
	Reason    string  `json:"reason"`
}

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

func (p *Payment) IsCompleted() bool {
	return p.Status == StatusCompleted
}

func (p *Payment) IsFailed() bool {
	return p.Status == StatusFailed
}

func (p *Payment) IsRefunded() bool {
	return p.Status == StatusRefunded
}

func (p *Payment) CanBeRefunded() bool {
	return p.Status == StatusCompleted
}

func (p *Payment) Complete(transactionID string) {
	p.Status = StatusCompleted
	p.TransactionID = transactionID
	now := time.Now()
	p.ProcessedAt = &now
	p.UpdatedAt = now
}

func (p *Payment) Fail(reason string) {
	p.Status = StatusFailed
	p.FailureReason = reason
	now := time.Now()
	p.ProcessedAt = &now
	p.UpdatedAt = now
}

func (p *Payment) Refund() error {
	if !p.CanBeRefunded() {
		return NewCannotRefundError(p.ID, p.Status)
	}
	p.Status = StatusRefunded
	p.UpdatedAt = time.Now()
	return nil
}

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