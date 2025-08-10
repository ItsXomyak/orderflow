package notification

import "time"

type Type string

const (
	TypeOrderConfirmed Type = "order_confirmed"
	TypeOrderFailed    Type = "order_failed"
	TypeOrderCancelled Type = "order_cancelled"
	TypePaymentFailed  Type = "payment_failed"
)

// Channel канал доставки уведомления
type Channel string

const (
	ChannelEmail Channel = "email"
	ChannelSMS   Channel = "sms"
	ChannelPush  Channel = "push"
)

// Status статус уведомления
type Status string

const (
	StatusPending Status = "pending"
	StatusSent    Status = "sent"
	StatusFailed  Status = "failed"
)

// Notification представляет уведомление
type Notification struct {
	ID         string            `json:"id"`
	CustomerID string            `json:"customer_id"`
	OrderID    string            `json:"order_id"`
	Type       Type              `json:"type"`
	Channel    Channel           `json:"channel"`
	Status     Status            `json:"status"`
	Subject    string            `json:"subject"`
	Message    string            `json:"message"`
	Metadata   map[string]string `json:"metadata,omitempty"`
	SentAt     *time.Time        `json:"sent_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

// Request запрос на отправку уведомления
type Request struct {
	CustomerID string            `json:"customer_id"`
	OrderID    string            `json:"order_id"`
	Type       Type              `json:"type"`
	Channel    Channel           `json:"channel"`
	Subject    string            `json:"subject,omitempty"`
	Message    string            `json:"message"`
	Metadata   map[string]string `json:"metadata,omitempty"`
}

// NewNotification создает новое уведомление
func NewNotification(req *Request) *Notification {
	return &Notification{
		CustomerID: req.CustomerID,
		OrderID:    req.OrderID,
		Type:       req.Type,
		Channel:    req.Channel,
		Subject:    req.Subject,
		Message:    req.Message,
		Metadata:   req.Metadata,
		Status:     StatusPending,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}
}

// IsSent проверяет, отправлено ли уведомление
func (n *Notification) IsSent() bool {
	return n.Status == StatusSent
}

// IsFailed проверяет, провалилась ли отправка
func (n *Notification) IsFailed() bool {
	return n.Status == StatusFailed
}

// MarkAsSent помечает уведомление как отправленное
func (n *Notification) MarkAsSent() {
	n.Status = StatusSent
	now := time.Now()
	n.SentAt = &now
	n.UpdatedAt = now
}

// MarkAsFailed помечает уведомление как неудачное
func (n *Notification) MarkAsFailed() {
	n.Status = StatusFailed
	n.UpdatedAt = time.Now()
}

// Validate валидирует уведомление
func (n *Notification) Validate() error {
	if n.CustomerID == "" {
		return NewValidationError("customer_id is required")
	}
	if n.OrderID == "" {
		return NewValidationError("order_id is required")
	}
	if n.Type == "" {
		return NewValidationError("type is required")
	}
	if n.Channel == "" {
		return NewValidationError("channel is required")
	}
	if n.Message == "" {
		return NewValidationError("message is required")
	}
	return nil
}