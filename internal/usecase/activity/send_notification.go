// internal/usecase/activity/send_notification.go
package activity

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"

	"orderflow/internal/domain/notification"
	"orderflow/internal/domain/order"
	"orderflow/internal/domain/workflow"
)

type SendNotificationActivity struct {
	notificationService notification.Service
	orderService        order.Service
}

func NewSendNotificationActivity(notificationService notification.Service, orderService order.Service) *SendNotificationActivity {
	return &SendNotificationActivity{
		notificationService: notificationService,
		orderService:        orderService,
	}
}

// отправка уведомления
func (a *SendNotificationActivity) Execute(ctx context.Context, input *workflow.SendNotificationActivityInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SendNotificationActivity", 
		"order_id", input.OrderID, 
		"customer_id", input.CustomerID,
		"type", input.Type)

	// валидация входных данных
	if err := input.Validate(); err != nil {
		logger.Error("Validation failed", "error", err)
		return workflow.NewActivityError(
			workflow.SendNotificationActivity,
			workflow.StepSendNotification,
			workflow.ErrorCodeValidation,
			err.Error(),
			false,
		)
	}

	// симуляция времени отправки уведомления
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(workflow.NotificationDuration):
	}

	// запрос на отправку уведомления
	notificationReq := &notification.Request{
		CustomerID: input.CustomerID,
		OrderID:    input.OrderID,
		Type:       input.Type,
		Channel:    input.Channel,
		Message:    input.Message,
		Metadata: map[string]string{
			"order_id": input.OrderID,
		},
	}

	// Если сообщение не указано, генерируем стандартное
	if notificationReq.Message == "" {
		notificationReq.Message = a.generateDefaultMessage(input.Type, input.OrderID)
	}

	// отправка уведомления
	if err := a.notificationService.Send(ctx, notificationReq); err != nil {
		logger.Error("Failed to send notification", "error", err)
		
		//  можно ли повторить операцию
		retryable := true
		errorCode := workflow.ErrorCodeNotificationFailed
		
		switch err.(type) {
		case *notification.ValidationError:
			retryable = false
			errorCode = workflow.ErrorCodeValidation
		case *notification.UnsupportedChannelError:
			retryable = false
		}
		
		return workflow.NewActivityError(
			workflow.SendNotificationActivity,
			workflow.StepSendNotification,
			errorCode,
			err.Error(),
			retryable,
		)
	}

	logger.Info("Notification sent successfully", 
		"order_id", input.OrderID,
		"type", input.Type,
		"channel", input.Channel)

	return nil
}

// отправка уведомления о подтверждении заказа
func (a *SendNotificationActivity) SendOrderConfirmation(ctx context.Context, orderID, customerID, paymentID string) error {
	input := &workflow.SendNotificationActivityInput{
		CustomerID: customerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderConfirmed,
		Channel:    notification.ChannelEmail,
		Message:    a.generateOrderConfirmationMessage(orderID, paymentID),
	}
	
	return a.Execute(ctx, input)
}

// отправка уведомления о неудачном заказе
func (a *SendNotificationActivity) SendOrderFailure(ctx context.Context, orderID, customerID, reason string) error {
	input := &workflow.SendNotificationActivityInput{
		CustomerID: customerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderFailed,
		Channel:    notification.ChannelEmail,
		Message:    a.generateOrderFailureMessage(orderID, reason),
	}
	
	return a.Execute(ctx, input)
}

// отправка уведомления об отмене заказа
func (a *SendNotificationActivity) SendOrderCancellation(ctx context.Context, orderID, customerID string) error {
	input := &workflow.SendNotificationActivityInput{
		CustomerID: customerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderCancelled,
		Channel:    notification.ChannelEmail,
		Message:    a.generateOrderCancellationMessage(orderID),
	}
	
	return a.Execute(ctx, input)
}

// генерирует стандартное сообщение для типа уведомления
func (a *SendNotificationActivity) generateDefaultMessage(notificationType notification.Type, orderID string) string {
	switch notificationType {
	case notification.TypeOrderConfirmed:
		return a.generateOrderConfirmationMessage(orderID, "")
	case notification.TypeOrderFailed:
		return a.generateOrderFailureMessage(orderID, "Processing failed")
	case notification.TypeOrderCancelled:
		return a.generateOrderCancellationMessage(orderID)
	default:
		return "Order update for order " + orderID
	}
}

// генерирует сообщение о подтверждении заказа
func (a *SendNotificationActivity) generateOrderConfirmationMessage(orderID, paymentID string) string {
	message := "Your order " + orderID + " has been successfully processed"
	if paymentID != "" {
		message += " (Payment ID: " + paymentID + ")"
	}
	message += ". Thank you for your purchase!"
	return message
}

// генерирует сообщение о неудачном заказе
func (a *SendNotificationActivity) generateOrderFailureMessage(orderID, reason string) string {
	message := "Unfortunately, your order " + orderID + " could not be processed"
	if reason != "" {
		message += ": " + reason
	}
	message += ". Please contact our support team for assistance."
	return message
}

// генерирует сообщение об отмене заказа
func (a *SendNotificationActivity) generateOrderCancellationMessage(orderID string) string {
	return "Your order " + orderID + " has been cancelled as requested. If you have any questions, please contact our support team."
}

func (a *SendNotificationActivity) GetActivityName() string {
	return workflow.SendNotificationActivity
}