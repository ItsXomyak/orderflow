package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"orderflow/internal/domain/notification"
	"orderflow/pkg/logger"
)

type NotificationService struct {
	notificationRepo notification.Repository
	senders map[notification.Channel]notification.Sender
	template notification.Template
}

func NewNotificationService(notificationRepo notification.Repository) *NotificationService {
	service := &NotificationService{
		notificationRepo: notificationRepo,
		senders:          make(map[notification.Channel]notification.Sender),
		template:         NewNotificationTemplate(),
	}
	
	service.initializeSenders()
	
	return service
}

func (service *NotificationService) initializeSenders() {
	service.senders[notification.ChannelEmail] = NewEmailSender()
	service.senders[notification.ChannelSMS] = NewSMSSender()
	service.senders[notification.ChannelPush] = NewPushSender()
}

func (service *NotificationService) Send(ctx context.Context, req *notification.Request) error {
	logger.Info("Sending notification", 
		"order_id", req.OrderID, 
		"customer_id", req.CustomerID,
		"type", req.Type,
		"channel", req.Channel)

	if req.CustomerID == "" {
		return notification.NewValidationError("customer_id is required")
	}
	if req.OrderID == "" {
		return notification.NewValidationError("order_id is required")
	}
	if req.Type == "" {
		return notification.NewValidationError("type is required")
	}
	if req.Channel == "" {
		return notification.NewValidationError("channel is required")
	}

	notificationEntity := notification.NewNotification(req)
	notificationEntity.ID = uuid.New().String()

	if notificationEntity.Message == "" {
		subject, message, err := service.template.Render(ctx, req.Type, map[string]interface{}{
			"order_id":    req.OrderID,
			"customer_id": req.CustomerID,
		})
		if err != nil {
			logger.Error("Failed to render notification template", "error", err)
			notificationEntity.MarkAsFailed()
		} else {
			notificationEntity.Subject = subject
			notificationEntity.Message = message
		}
	}

	if err := service.notificationRepo.CreateNotification(ctx, notificationEntity); err != nil {
		return err
	}

	_, exists := service.senders[req.Channel]
	if !exists {
		logger.Error("Unsupported notification channel", "channel", req.Channel)
		notificationEntity.MarkAsFailed()
		service.notificationRepo.UpdateNotification(ctx, notificationEntity)
		return notification.NewUnsupportedChannelError(req.Channel)
	}

	success := service.simulateNotificationSending(req.Channel)
	if success {
		notificationEntity.MarkAsSent()
		logger.Info("Notification sent successfully", 
			"notification_id", notificationEntity.ID,
			"order_id", req.OrderID,
			"channel", req.Channel)
	} else {
		notificationEntity.MarkAsFailed()
		logger.Error("Failed to send notification", 
			"notification_id", notificationEntity.ID,
			"order_id", req.OrderID,
			"channel", req.Channel)
	}

	if err := service.notificationRepo.UpdateNotification(ctx, notificationEntity); err != nil {
		return err
	}

	if !success {
		return notification.NewSendError(req.Channel, "Failed to send notification")
	}

	return nil
}

func (service *NotificationService) simulateNotificationSending(channel notification.Channel) bool {
	successRates := map[notification.Channel]int{
		notification.ChannelEmail: 95, // 95% успешных email
		notification.ChannelSMS:   90, // 90% успешных SMS
		notification.ChannelPush:  85, // 85% успешных push
	}

	rate, exists := successRates[channel]
	if !exists {
		rate = 80 // по умолчанию 80%
	}

	return time.Now().UnixNano()%100 < int64(rate)
}

func (service *NotificationService) GetByID(ctx context.Context, id string) (*notification.Notification, error) {
	if id == "" {
		return nil, notification.NewValidationError("notification_id is required")
	}

	return service.notificationRepo.GetNotification(ctx, id)
}

func (service *NotificationService) GetByOrderID(ctx context.Context, orderID string) ([]*notification.Notification, error) {
	if orderID == "" {
		return nil, notification.NewValidationError("order_id is required")
	}

	return service.notificationRepo.GetNotificationsByOrderID(ctx, orderID)
}

func (service *NotificationService) Retry(ctx context.Context, id string) error {
	logger.Info("Retrying notification", "notification_id", id)

	if id == "" {
		return notification.NewValidationError("notification_id is required")
	}

	notificationEntity, err := service.notificationRepo.GetNotification(ctx, id)
	if err != nil {
		return err
	}

	if !notificationEntity.IsFailed() {
		return notification.NewValidationError("Notification is not in failed status")
	}

	_, exists := service.senders[notificationEntity.Channel]
	if !exists {
		return notification.NewUnsupportedChannelError(notificationEntity.Channel)
	}

	success := service.simulateNotificationSending(notificationEntity.Channel)
	if success {
		notificationEntity.MarkAsSent()
		logger.Info("Notification retry successful", "notification_id", id)
	} else {
		notificationEntity.MarkAsFailed()
		logger.Error("Notification retry failed", "notification_id", id)
		return notification.NewSendError(notificationEntity.Channel, "Retry failed")
	}

	if err := service.notificationRepo.UpdateNotification(ctx, notificationEntity); err != nil {
		return err
	}

	return nil
}

func (service *NotificationService) GetNotifications(ctx context.Context) ([]*notification.Notification, error) {
	return service.notificationRepo.GetNotifications(ctx)
}

func (service *NotificationService) GetNotificationStatistics(ctx context.Context) (*NotificationStatistics, error) {
	notifications, err := service.notificationRepo.GetNotifications(ctx)
	if err != nil {
		return nil, err
	}

	stats := &NotificationStatistics{
		TotalNotifications: 0,
		SentNotifications:  0,
		FailedNotifications: 0,
		PendingNotifications: 0,
		ChannelStats:        make(map[notification.Channel]ChannelStats),
	}

	for _, notificationEntity := range notifications {
		stats.TotalNotifications++

		switch notificationEntity.Status {
		case notification.StatusSent:
			stats.SentNotifications++
		case notification.StatusFailed:
			stats.FailedNotifications++
		case notification.StatusPending:
			stats.PendingNotifications++
		}

		channelStats := stats.ChannelStats[notificationEntity.Channel]
		channelStats.Total++
		switch notificationEntity.Status {
		case notification.StatusSent:
			channelStats.Sent++
		case notification.StatusFailed:
			channelStats.Failed++
		case notification.StatusPending:
			channelStats.Pending++
		}
		stats.ChannelStats[notificationEntity.Channel] = channelStats
	}

	if stats.TotalNotifications > 0 {
		stats.SuccessRate = float64(stats.SentNotifications) / float64(stats.TotalNotifications) * 100
	}

	return stats, nil
}

func (service *NotificationService) RetryFailedNotifications(ctx context.Context) error {
	logger.Info("Retrying all failed notifications")

	failedNotifications, err := service.notificationRepo.GetFailedNotifications(ctx)
	if err != nil {
		return err
	}

	successCount := 0
	for _, notificationEntity := range failedNotifications {
		if err := service.Retry(ctx, notificationEntity.ID); err != nil {
			logger.Error("Failed to retry notification", 
				"notification_id", notificationEntity.ID, 
				"error", err)
		} else {
			successCount++
		}
	}

	logger.Info("Failed notifications retry completed", 
		"total", len(failedNotifications), 
		"successful", successCount)
	return nil
}

type NotificationStatistics struct {
	TotalNotifications  int                                    `json:"total_notifications"`
	SentNotifications   int                                    `json:"sent_notifications"`
	FailedNotifications int                                    `json:"failed_notifications"`
	PendingNotifications int                                   `json:"pending_notifications"`
	SuccessRate         float64                                `json:"success_rate"`
	ChannelStats        map[notification.Channel]ChannelStats `json:"channel_stats"`
}

type ChannelStats struct {
	Total   int `json:"total"`
	Sent    int `json:"sent"`
	Failed  int `json:"failed"`
	Pending int `json:"pending"`
}

type EmailSender struct{}

func NewEmailSender() *EmailSender {
	return &EmailSender{}
}

func (s *EmailSender) Send(ctx context.Context, notification *notification.Notification) error {
	logger.Info("Sending email notification", 
		"to", notification.CustomerID,
		"subject", notification.Subject)
	return nil
}

func (s *EmailSender) SupportedChannels() []notification.Channel {
	return []notification.Channel{notification.ChannelEmail}
}

type SMSSender struct{}

func NewSMSSender() *SMSSender {
	return &SMSSender{}
}

func (s *SMSSender) Send(ctx context.Context, notification *notification.Notification) error {
	logger.Info("Sending SMS notification", 
		"to", notification.CustomerID,
		"message", notification.Message)
	return nil
}

func (s *SMSSender) SupportedChannels() []notification.Channel {
	return []notification.Channel{notification.ChannelSMS}
}

type PushSender struct{}

func NewPushSender() *PushSender {
	return &PushSender{}
}

func (s *PushSender) Send(ctx context.Context, notification *notification.Notification) error {
	logger.Info("Sending push notification", 
		"to", notification.CustomerID,
		"title", notification.Subject,
		"body", notification.Message)
	return nil
}

func (s *PushSender) SupportedChannels() []notification.Channel {
	return []notification.Channel{notification.ChannelPush}
}

type NotificationTemplate struct{}

func NewNotificationTemplate() *NotificationTemplate {
	return &NotificationTemplate{}
}

func (t *NotificationTemplate) Render(ctx context.Context, notificationType notification.Type, data map[string]interface{}) (string, string, error) {
	subject, err := t.GetSubject(ctx, notificationType, data)
	if err != nil {
		return "", "", err
	}

	message, err := t.GetMessage(ctx, notificationType, data)
	if err != nil {
		return "", "", err
	}

	return subject, message, nil
}

func (t *NotificationTemplate) GetSubject(ctx context.Context, notificationType notification.Type, data map[string]interface{}) (string, error) {
	orderID, _ := data["order_id"].(string)
	
	switch notificationType {
	case notification.TypeOrderConfirmed:
		return fmt.Sprintf("Order Confirmed - %s", orderID), nil
	case notification.TypeOrderFailed:
		return fmt.Sprintf("Order Failed - %s", orderID), nil
	case notification.TypeOrderCancelled:
		return fmt.Sprintf("Order Cancelled - %s", orderID), nil
	case notification.TypePaymentFailed:
		return fmt.Sprintf("Payment Failed - %s", orderID), nil
	default:
		return "Order Update", nil
	}
}

func (t *NotificationTemplate) GetMessage(ctx context.Context, notificationType notification.Type, data map[string]interface{}) (string, error) {
	orderID, _ := data["order_id"].(string)
	
	switch notificationType {
	case notification.TypeOrderConfirmed:
		return fmt.Sprintf("Your order %s has been successfully processed and confirmed. Thank you for your purchase!", orderID), nil
	case notification.TypeOrderFailed:
		return fmt.Sprintf("Unfortunately, your order %s could not be processed. Please contact our support team for assistance.", orderID), nil
	case notification.TypeOrderCancelled:
		return fmt.Sprintf("Your order %s has been cancelled as requested. If you have any questions, please contact our support team.", orderID), nil
	case notification.TypePaymentFailed:
		return fmt.Sprintf("Payment for your order %s has failed. Please check your payment method and try again.", orderID), nil
	default:
		return fmt.Sprintf("There has been an update to your order %s.", orderID), nil
	}
}
