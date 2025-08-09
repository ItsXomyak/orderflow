package notification

import (
	"fmt"
	"time"
)

type NotificationType string

const (
	NotificationTypeEmail   NotificationType = "email"
	NotificationTypeSMS     NotificationType = "sms"
	NotificationTypePush    NotificationType = "push"
)

type NotificationStatus string

const (
	NotificationStatusPending   NotificationStatus = "pending"
	NotificationStatusSending   NotificationStatus = "sending"
	NotificationStatusSent      NotificationStatus = "sent"
	NotificationStatusFailed    NotificationStatus = "failed"
)

type Notification struct {
	ID        string             `json:"id"`
	OrderID   string             `json:"order_id"`
	CustomerID string            `json:"customer_id"`
	Type      NotificationType   `json:"type"`
	Subject   string             `json:"subject"`
	Message   string             `json:"message"`
	Status    NotificationStatus `json:"status"`
	SentAt    *time.Time         `json:"sent_at,omitempty"`
	CreatedAt time.Time          `json:"created_at"`
	Error     string             `json:"error,omitempty"`
}

type NotificationRequest struct {
	OrderID    string            `json:"order_id"`
	CustomerID string            `json:"customer_id"`
	Type       NotificationType  `json:"type"`
	Subject    string            `json:"subject"`
	Message    string            `json:"message"`
}

type NotificationResponse struct {
	NotificationID string             `json:"notification_id"`
	OrderID        string             `json:"order_id"`
	Status         NotificationStatus `json:"status"`
	Message        string             `json:"message"`
}

type MockNotificationService struct {
	notifications map[string]*Notification
}

func NewMockNotificationService() *MockNotificationService {
	return &MockNotificationService{
		notifications: make(map[string]*Notification),
	}
}

func (s *MockNotificationService) SendNotification(request NotificationRequest) (*NotificationResponse, error) {
	notification := &Notification{
		ID:         fmt.Sprintf("NOTIF-%s-%d", request.OrderID, time.Now().Unix()),
		OrderID:    request.OrderID,
		CustomerID: request.CustomerID,
		Type:       request.Type,
		Subject:    request.Subject,
		Message:    request.Message,
		Status:     NotificationStatusSending,
		CreatedAt:  time.Now(),
	}

	time.Sleep(50 * time.Millisecond)

	switch request.Type {
	case NotificationTypeEmail:
		if s.shouldFailEmail() {
			notification.Status = NotificationStatusFailed
			notification.Error = "Email service temporarily unavailable"
			return &NotificationResponse{
				NotificationID: notification.ID,
				OrderID:        request.OrderID,
				Status:         NotificationStatusFailed,
				Message:        notification.Error,
			}, fmt.Errorf("notification failed: %s", notification.Error)
		}
	case NotificationTypeSMS:
		if s.shouldFailSMS() {
			notification.Status = NotificationStatusFailed
			notification.Error = "SMS gateway error"
			return &NotificationResponse{
				NotificationID: notification.ID,
				OrderID:        request.OrderID,
				Status:         NotificationStatusFailed,
				Message:        notification.Error,
			}, fmt.Errorf("notification failed: %s", notification.Error)
		}
	}

	notification.Status = NotificationStatusSent
	now := time.Now()
	notification.SentAt = &now

	s.notifications[notification.ID] = notification

	return &NotificationResponse{
		NotificationID: notification.ID,
		OrderID:        request.OrderID,
		Status:         NotificationStatusSent,
		Message:        "Notification sent successfully",
	}, nil
}

func (s *MockNotificationService) shouldFailEmail() bool {
	return time.Now().UnixNano()%10 == 0
}

func (s *MockNotificationService) shouldFailSMS() bool {
	return time.Now().UnixNano()%20 == 0
}

func (s *MockNotificationService) GetNotification(notificationID string) (*Notification, error) {
	notification, exists := s.notifications[notificationID]
	if !exists {
		return nil, fmt.Errorf("notification %s not found", notificationID)
	}
	return notification, nil
}
