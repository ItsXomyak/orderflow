package workflow

import (
	"context"
	"fmt"
	"time"

	"go.temporal.io/sdk/activity"
	"go.uber.org/zap"

	"order-processing/internal/domain/inventory"
	"order-processing/internal/domain/notification"
	"order-processing/internal/domain/payment"
)

type CreateOrderResult struct {
	OrderID string `json:"order_id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

type InventoryCheckResult struct {
	OrderID   string `json:"order_id"`
	Available bool   `json:"available"`
	Message   string `json:"message"`
}

type PaymentResult struct {
	PaymentID     string `json:"payment_id"`
	OrderID       string `json:"order_id"`
	Status        string `json:"status"`
	TransactionID string `json:"transaction_id,omitempty"`
	Message       string `json:"message"`
}

type NotificationResult struct {
	NotificationID string `json:"notification_id"`
	OrderID        string `json:"order_id"`
	Status         string `json:"status"`
	Message        string `json:"message"`
}


func CreateOrderActivity(ctx context.Context, input OrderWorkflowInput) (*CreateOrderResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Creating order", 
		zap.String("orderID", input.OrderID),
		zap.String("customerID", input.CustomerID),
		zap.Float64("totalAmount", input.TotalAmount),
	)

	time.Sleep(100 * time.Millisecond)


	logger.Info("Order created successfully", 
		zap.String("orderID", input.OrderID),
	)

	return &CreateOrderResult{
		OrderID: input.OrderID,
		Status:  "created",
		Message: "Order created successfully",
	}, nil
}

func CheckInventoryActivity(ctx context.Context, input OrderWorkflowInput) (*InventoryCheckResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Checking inventory", 
		zap.String("orderID", input.OrderID),
		zap.Int("itemCount", len(input.Items)),
	)

	inventoryService := inventory.NewMockInventoryService()

	for _, item := range input.Items {
		logger.Info("Checking item availability", 
			zap.String("productID", item.ProductID),
			zap.Int("quantity", item.Quantity),
		)

		request := inventory.InventoryCheckRequest{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}

		response, err := inventoryService.CheckAvailability(request)
		if err != nil {
			logger.Error("Failed to check inventory for item", 
				zap.String("productID", item.ProductID),
				zap.Error(err),
			)
			return &InventoryCheckResult{
				OrderID:   input.OrderID,
				Available: false,
				Message:   fmt.Sprintf("Failed to check inventory for product %s: %v", item.ProductID, err),
			}, err
		}

		if !response.Available {
			logger.Warn("Item not available", 
				zap.String("productID", item.ProductID),
				zap.String("message", response.Message),
			)
			return &InventoryCheckResult{
				OrderID:   input.OrderID,
				Available: false,
				Message:   response.Message,
			}, nil
		}

		logger.Info("Item is available", 
			zap.String("productID", item.ProductID),
			zap.Int("availableQty", response.AvailableQty),
		)
	}

	logger.Info("All items are available", zap.String("orderID", input.OrderID))

	return &InventoryCheckResult{
		OrderID:   input.OrderID,
		Available: true,
		Message:   "All items are available",
	}, nil
}

func ProcessPaymentActivity(ctx context.Context, input OrderWorkflowInput) (*PaymentResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Processing payment", 
		zap.String("orderID", input.OrderID),
		zap.Float64("amount", input.TotalAmount),
	)

	paymentService := payment.NewMockPaymentService()

	paymentRequest := payment.PaymentRequest{
		OrderID:    input.OrderID,
		CustomerID: input.CustomerID,
		Amount:     input.TotalAmount,
		Currency:   "USD",
		Method:     payment.PaymentMethodCreditCard,
	}

	paymentResponse, err := paymentService.ProcessPayment(paymentRequest)
	if err != nil {
		logger.Error("Payment processing failed", 
			zap.String("orderID", input.OrderID),
			zap.Error(err),
		)
		return &PaymentResult{
			PaymentID: "",
			OrderID:   input.OrderID,
			Status:    "failed",
			Message:   fmt.Sprintf("Payment processing failed: %v", err),
		}, err
	}

	logger.Info("Payment processed successfully", 
		zap.String("orderID", input.OrderID),
		zap.String("paymentID", paymentResponse.PaymentID),
		zap.String("transactionID", paymentResponse.TransactionID),
	)

	return &PaymentResult{
		PaymentID:     paymentResponse.PaymentID,
		OrderID:       paymentResponse.OrderID,
		Status:        string(paymentResponse.Status),
		TransactionID: paymentResponse.TransactionID,
		Message:       paymentResponse.Message,
	}, nil
}

func SendNotificationActivity(ctx context.Context, input OrderWorkflowInput, paymentResult PaymentResult) (*NotificationResult, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Sending notification", 
		zap.String("orderID", input.OrderID),
		zap.String("customerID", input.CustomerID),
	)

	notificationService := notification.NewMockNotificationService()

	notificationRequest := notification.NotificationRequest{
		OrderID:    input.OrderID,
		CustomerID: input.CustomerID,
		Type:       notification.NotificationTypeEmail,
		Subject:    fmt.Sprintf("Order Confirmation - %s", input.OrderID),
		Message:    fmt.Sprintf("Your order %s has been successfully processed. Total amount: $%.2f. Transaction ID: %s", 
			input.OrderID, input.TotalAmount, paymentResult.TransactionID),
	}

	notificationResponse, err := notificationService.SendNotification(notificationRequest)
	if err != nil {
		logger.Error("Failed to send notification", 
			zap.String("orderID", input.OrderID),
			zap.Error(err),
		)
		return &NotificationResult{
			NotificationID: "",
			OrderID:        input.OrderID,
			Status:         "failed",
			Message:        fmt.Sprintf("Failed to send notification: %v", err),
		}, err
	}

	logger.Info("Notification sent successfully", 
		zap.String("orderID", input.OrderID),
		zap.String("notificationID", notificationResponse.NotificationID),
	)

	return &NotificationResult{
		NotificationID: notificationResponse.NotificationID,
		OrderID:        notificationResponse.OrderID,
		Status:         string(notificationResponse.Status),
		Message:        notificationResponse.Message,
	}, nil
}

func CancelOrderActivity(ctx context.Context, orderID string) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Cancelling order", zap.String("orderID", orderID))

	time.Sleep(50 * time.Millisecond)

	logger.Info("Order cancelled successfully", zap.String("orderID", orderID))

	return nil
}
