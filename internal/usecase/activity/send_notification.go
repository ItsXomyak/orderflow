package activity

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"

	"orderflow/internal/domain/notification"
	"orderflow/internal/domain/order"
	wf "orderflow/internal/domain/workflow"
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

func (a *SendNotificationActivity) Execute(ctx context.Context, input *wf.SendNotificationActivityInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting SendNotificationActivity", 
		"order_id", input.OrderID, 
		"customer_id", input.CustomerID,
		"type", input.Type)

	if err := input.Validate(); err != nil {
		logger.Error("Validation failed", "error", err)
		return wf.NewActivityError(
			wf.SendNotificationActivity,
			wf.StepSendNotification,
			wf.ErrorCodeValidation,
			err.Error(),
			false,
		)
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wf.NotificationDuration):
	}

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

	if notificationReq.Message == "" {
		notificationReq.Message = a.generateDefaultMessage(input.Type, input.OrderID)
	}

	if err := a.notificationService.Send(ctx, notificationReq); err != nil {
		logger.Error("Failed to send notification", "error", err)
		
		retryable := true
		errorCode := wf.ErrorCodeNotificationFailed
		
		switch err.(type) {
		case *notification.ValidationError:
			retryable = false
			errorCode = wf.ErrorCodeValidation
		case *notification.UnsupportedChannelError:
			retryable = false
		}
		
		return wf.NewActivityError(
			wf.SendNotificationActivity,
			wf.StepSendNotification,
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

func (a *SendNotificationActivity) SendOrderConfirmation(ctx context.Context, orderID, customerID, paymentID string) error {
	input := &wf.SendNotificationActivityInput{
		CustomerID: customerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderConfirmed,
		Channel:    notification.ChannelEmail,
		Message:    a.generateOrderConfirmationMessage(orderID, paymentID),
	}
	
	return a.Execute(ctx, input)
}

func (a *SendNotificationActivity) SendOrderFailure(ctx context.Context, orderID, customerID, reason string) error {
	input := &wf.SendNotificationActivityInput{
		CustomerID: customerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderFailed,
		Channel:    notification.ChannelEmail,
		Message:    a.generateOrderFailureMessage(orderID, reason),
	}
	
	return a.Execute(ctx, input)
}

func (a *SendNotificationActivity) SendOrderCancellation(ctx context.Context, orderID, customerID string) error {
	input := &wf.SendNotificationActivityInput{
		CustomerID: customerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderCancelled,
		Channel:    notification.ChannelEmail,
		Message:    a.generateOrderCancellationMessage(orderID),
	}
	
	return a.Execute(ctx, input)
}

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

func (a *SendNotificationActivity) generateOrderConfirmationMessage(orderID, paymentID string) string {
	message := "Your order " + orderID + " has been successfully processed"
	if paymentID != "" {
		message += " (Payment ID: " + paymentID + ")"
	}
	message += ". Thank you for your purchase!"
	return message
}

func (a *SendNotificationActivity) generateOrderFailureMessage(orderID, reason string) string {
	message := "Unfortunately, your order " + orderID + " could not be processed"
	if reason != "" {
		message += ": " + reason
	}
	message += ". Please contact our support team for assistance."
	return message
}

func (a *SendNotificationActivity) generateOrderCancellationMessage(orderID string) string {
	return "Your order " + orderID + " has been cancelled as requested. If you have any questions, please contact our support team."
}

func (a *SendNotificationActivity) GetActivityName() (string, error) {
	return wf.SendNotificationActivity, nil
}