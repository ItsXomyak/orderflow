package order

import (
	"time"
)

type Status string

const (
	StatusPending    Status = "pending"
	StatusValidating Status = "validating"
	StatusPayment    Status = "payment"
	StatusCompleted  Status = "completed"
	StatusFailed     Status = "failed"
	StatusCancelled  Status = "cancelled"
)

type Order struct {
	ID          string    `json:"id"`
	CustomerID  string    `json:"customer_id"`
	Items       []Item    `json:"items"`
	TotalAmount float64   `json:"total_amount"`
	Status      Status    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`

	PaymentID     string     `json:"payment_id,omitempty"`
	FailureReason string     `json:"failure_reason,omitempty"`
	CompletedAt   *time.Time `json:"completed_at,omitempty"`
}

type Item struct {
	ProductID string  `json:"product_id"`
	Name      string  `json:"name"`
	Quantity  int     `json:"quantity"`
	Price     float64 `json:"price"`
}

type CreateRequest struct {
	CustomerID string `json:"customer_id"`
	Items      []Item `json:"items"`
}

func NewOrder(customerID string, items []Item) *Order {
	order := &Order{
		CustomerID: customerID,
		Items:      items,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
	order.CalculateTotal()
	return order
}

func (o *Order) IsCompleted() bool {
	return o.Status == StatusCompleted
}

func (o *Order) IsFailed() bool {
	return o.Status == StatusFailed
}

func (o *Order) IsCancelled() bool {
	return o.Status == StatusCancelled
}

func (o *Order) CanBeCancelled() bool {
	return o.Status == StatusPending ||
		o.Status == StatusValidating ||
		o.Status == StatusPayment
}

func (o *Order) UpdateStatus(status Status) {
	o.Status = status
	o.UpdatedAt = time.Now()

	if status == StatusCompleted {
		now := time.Now()
		o.CompletedAt = &now
	}
}

func (o *Order) SetFailure(reason string) {
	o.Status = StatusFailed
	o.FailureReason = reason
	o.UpdatedAt = time.Now()
}

func (o *Order) Cancel() error {
	if !o.CanBeCancelled() {
		return NewCannotCancelError(o.Status)
	}
	o.Status = StatusCancelled
	o.UpdatedAt = time.Now()
	return nil
}

func (o *Order) CalculateTotal() float64 {
	total := 0.0
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	o.TotalAmount = total
	return total
}

func (o *Order) Validate() error {
	if o.CustomerID == "" {
		return NewValidationError("customer_id is required")
	}

	if len(o.Items) == 0 {
		return NewValidationError("order must have at least one item")
	}

	for _, item := range o.Items {
		if item.ProductID == "" {
			return NewValidationError("product_id is required for all items")
		}
		if item.Quantity <= 0 {
			return NewValidationError("quantity must be positive")
		}
		if item.Price < 0 {
			return NewValidationError("price cannot be negative")
		}
	}

	return nil
}
