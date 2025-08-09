package order

import (
	"time"

	"github.com/google/uuid"
)

type OrderStatus string

const (
	OrderStatusCreated    OrderStatus = "created"
	OrderStatusValidating OrderStatus = "validating"
	OrderStatusProcessing OrderStatus = "processing"
	OrderStatusPaid       OrderStatus = "paid"
	OrderStatusCompleted  OrderStatus = "completed"
	OrderStatusFailed     OrderStatus = "failed"
	OrderStatusCancelled  OrderStatus = "cancelled"
)

type Order struct {
	ID          uuid.UUID   `json:"id"`
	CustomerID  string      `json:"customer_id"`
	Items       []OrderItem `json:"items"`
	TotalAmount float64     `json:"total_amount"`
	Status      OrderStatus `json:"status"`
	CreatedAt   time.Time   `json:"created_at"`
	UpdatedAt   time.Time   `json:"updated_at"`
	Error       string      `json:"error,omitempty"`
}

type OrderItem struct {
	ProductID   string  `json:"product_id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	UnitPrice   float64 `json:"unit_price"`
	TotalPrice  float64 `json:"total_price"`
}

type OrderRequest struct {
	CustomerID string      `json:"customer_id" binding:"required"`
	Items      []OrderItem `json:"items" binding:"required"`
}

func NewOrder(request OrderRequest) *Order {
	totalAmount := 0.0
	for i := range request.Items {
		request.Items[i].TotalPrice = float64(request.Items[i].Quantity) * request.Items[i].UnitPrice
		totalAmount += request.Items[i].TotalPrice
	}

	return &Order{
		ID:          uuid.New(),
		CustomerID:  request.CustomerID,
		Items:       request.Items,
		TotalAmount: totalAmount,
		Status:      OrderStatusCreated,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (o *Order) UpdateStatus(status OrderStatus) {
	o.Status = status
	o.UpdatedAt = time.Now()
}

func (o *Order) SetError(err string) {
	o.Error = err
	o.Status = OrderStatusFailed
	o.UpdatedAt = time.Now()
}

func (o *Order) CanCancel() bool {
	return o.Status == OrderStatusCreated || 
		   o.Status == OrderStatusValidating || 
		   o.Status == OrderStatusProcessing
}

func (o *Order) Cancel() bool {
	if !o.CanCancel() {
		return false
	}
	o.UpdateStatus(OrderStatusCancelled)
	return true
}
