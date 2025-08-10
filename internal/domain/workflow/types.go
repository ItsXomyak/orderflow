package workflow

import (
	"orderflow/internal/domain/inventory"
	"orderflow/internal/domain/notification"
	"orderflow/internal/domain/order"
)

type OrderProcessingInput struct {
	CustomerID string       `json:"customer_id"`
	Items      []order.Item `json:"items"`
}

type ActivityInput interface {
	Validate() error
}

type CreateOrderActivityInput struct {
	CustomerID string       `json:"customer_id"`
	Items      []order.Item `json:"items"`
}

func (i *CreateOrderActivityInput) Validate() error {
	if i.CustomerID == "" {
		return NewValidationError("customer_id is required")
	}
	if len(i.Items) == 0 {
		return NewValidationError("items are required")
	}
	return nil
}

type CreateOrderActivityOutput struct {
	OrderID string `json:"order_id"`
}

type CheckInventoryActivityInput struct {
	OrderID string       `json:"order_id"`
	Items   []order.Item `json:"items"`
}

func (i *CheckInventoryActivityInput) Validate() error {
	if i.OrderID == "" {
		return NewValidationError("order_id is required")
	}
	return nil
}

type CheckInventoryActivityOutput struct {
	Available        bool                            `json:"available"`
	UnavailableItems []inventory.UnavailableItem     `json:"unavailable_items,omitempty"`
}

type ProcessPaymentActivityInput struct {
	OrderID    string  `json:"order_id"`
	CustomerID string  `json:"customer_id"`
	Amount     float64 `json:"amount"`
	Currency   string  `json:"currency"`
}

func (i *ProcessPaymentActivityInput) Validate() error {
	if i.OrderID == "" {
		return NewValidationError("order_id is required")
	}
	if i.CustomerID == "" {
		return NewValidationError("customer_id is required")
	}
	if i.Amount <= 0 {
		return NewValidationError("amount must be positive")
	}
	return nil
}

type ProcessPaymentActivityOutput struct {
	PaymentID     string `json:"payment_id"`
	TransactionID string `json:"transaction_id"`
}

type SendNotificationActivityInput struct {
	CustomerID string               `json:"customer_id"`
	OrderID    string               `json:"order_id"`
	Type       notification.Type    `json:"type"`
	Channel    notification.Channel `json:"channel"`
	Message    string               `json:"message"`
}

func (i *SendNotificationActivityInput) Validate() error {
	if i.CustomerID == "" {
		return NewValidationError("customer_id is required")
	}
	if i.OrderID == "" {
		return NewValidationError("order_id is required")
	}
	return nil
}

type ActivityResult struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type WorkflowResult struct {
	OrderID   string      `json:"order_id"`
	Status    order.Status `json:"status"`
	Success   bool        `json:"success"`
	Message   string      `json:"message,omitempty"`
	PaymentID string      `json:"payment_id,omitempty"`
}