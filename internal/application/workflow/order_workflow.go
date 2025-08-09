package workflow

import (
	"fmt"
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"
	"go.uber.org/zap"

	"order-processing/internal/domain/order"
)

const (
	TaskQueueOrderProcessing = "order-processing"
	WorkflowIDPrefix         = "order-workflow"
)

type OrderWorkflowInput struct {
	OrderID    string             `json:"order_id"`
	CustomerID string             `json:"customer_id"`
	Items      []order.OrderItem  `json:"items"`
	TotalAmount float64           `json:"total_amount"`
}

type OrderWorkflowOutput struct {
	OrderID       string           `json:"order_id"`
	Status        order.OrderStatus `json:"status"`
	PaymentID     string           `json:"payment_id,omitempty"`
	TransactionID string           `json:"transaction_id,omitempty"`
	Error         string           `json:"error,omitempty"`
}

func OrderWorkflow(ctx workflow.Context, input OrderWorkflowInput) (*OrderWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	runID := workflow.GetInfo(ctx).WorkflowExecution.RunID

	logger.Info("Starting order workflow",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID),
		zap.String("orderID", input.OrderID),
		zap.String("customerID", input.CustomerID),
		zap.Float64("totalAmount", input.TotalAmount),
	)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        10 * time.Second,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"NonRetryableError"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	logger.Info("Step 1: Creating order", zap.String("orderID", input.OrderID))
	var createOrderResult CreateOrderResult
	err := workflow.ExecuteActivity(ctx, CreateOrderActivity, input).Get(ctx, &createOrderResult)
	if err != nil {
		logger.Error("Failed to create order", zap.Error(err), zap.String("orderID", input.OrderID))
		return &OrderWorkflowOutput{
			OrderID: input.OrderID,
			Status:  order.OrderStatusFailed,
			Error:   fmt.Sprintf("Failed to create order: %v", err),
		}, nil
	}

	logger.Info("Step 2: Checking inventory", zap.String("orderID", input.OrderID))
	var inventoryResult InventoryCheckResult
	err = workflow.ExecuteActivity(ctx, CheckInventoryActivity, input).Get(ctx, &inventoryResult)
	if err != nil {
		logger.Error("Failed to check inventory", zap.Error(err), zap.String("orderID", input.OrderID))
		return &OrderWorkflowOutput{
			OrderID: input.OrderID,
			Status:  order.OrderStatusFailed,
			Error:   fmt.Sprintf("Failed to check inventory: %v", err),
		}, nil
	}

	if !inventoryResult.Available {
		logger.Error("Inventory check failed", 
			zap.String("orderID", input.OrderID),
			zap.String("message", inventoryResult.Message),
		)
		return &OrderWorkflowOutput{
			OrderID: input.OrderID,
			Status:  order.OrderStatusFailed,
			Error:   inventoryResult.Message,
		}, nil
	}

	logger.Info("Step 3: Processing payment", zap.String("orderID", input.OrderID))
	var paymentResult PaymentResult
	err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, input).Get(ctx, &paymentResult)
	if err != nil {
		logger.Error("Failed to process payment", zap.Error(err), zap.String("orderID", input.OrderID))
		return &OrderWorkflowOutput{
			OrderID: input.OrderID,
			Status:  order.OrderStatusFailed,
			Error:   fmt.Sprintf("Failed to process payment: %v", err),
		}, nil
	}

	if paymentResult.Status != "completed" {
		logger.Error("Payment processing failed", 
			zap.String("orderID", input.OrderID),
			zap.String("status", paymentResult.Status),
			zap.String("message", paymentResult.Message),
		)
		return &OrderWorkflowOutput{
			OrderID: input.OrderID,
			Status:  order.OrderStatusFailed,
			Error:   paymentResult.Message,
		}, nil
	}

	logger.Info("Step 4: Sending notification", zap.String("orderID", input.OrderID))
	var notificationResult NotificationResult
	err = workflow.ExecuteActivity(ctx, SendNotificationActivity, input, paymentResult).Get(ctx, &notificationResult)
	if err != nil {
		logger.Error("Failed to send notification", zap.Error(err), zap.String("orderID", input.OrderID))
		logger.Warn("Notification failed, but order processing completed successfully", 
			zap.Error(err),
			zap.String("orderID", input.OrderID),
		)
	}

	logger.Info("Order workflow completed successfully", 
		zap.String("orderID", input.OrderID),
		zap.String("status", string(order.OrderStatusCompleted)),
	)

	return &OrderWorkflowOutput{
		OrderID:       input.OrderID,
		Status:        order.OrderStatusCompleted,
		PaymentID:     paymentResult.PaymentID,
		TransactionID: paymentResult.TransactionID,
	}, nil
}

func OrderWorkflowWithCancellation(ctx workflow.Context, input OrderWorkflowInput) (*OrderWorkflowOutput, error) {
	logger := workflow.GetLogger(ctx)
	workflowID := workflow.GetInfo(ctx).WorkflowExecution.ID
	runID := workflow.GetInfo(ctx).WorkflowExecution.RunID

	logger.Info("Starting order workflow with cancellation support",
		zap.String("workflowID", workflowID),
		zap.String("runID", runID),
		zap.String("orderID", input.OrderID),
	)

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 5 * time.Minute,
		RetryPolicy: &temporal.RetryPolicy{
			InitialInterval:        time.Second,
			BackoffCoefficient:     2.0,
			MaximumInterval:        10 * time.Second,
			MaximumAttempts:        3,
			NonRetryableErrorTypes: []string{"NonRetryableError"},
		},
	}
	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	cancelSignal := workflow.GetSignalChannel(ctx, "cancel-order")
	
	selector := workflow.NewSelector(ctx)

	resultCh := workflow.NewChannel(ctx)
	
	workflow.Go(ctx, func(ctx workflow.Context) {
		result := &OrderWorkflowOutput{}
		var err error

		logger.Info("Step 1: Creating order", zap.String("orderID", input.OrderID))
		var createOrderResult CreateOrderResult
		err = workflow.ExecuteActivity(ctx, CreateOrderActivity, input).Get(ctx, &createOrderResult)
		if err != nil {
			logger.Error("Failed to create order", zap.Error(err), zap.String("orderID", input.OrderID))
			result = &OrderWorkflowOutput{
				OrderID: input.OrderID,
				Status:  order.OrderStatusFailed,
				Error:   fmt.Sprintf("Failed to create order: %v", err),
			}
			resultCh.Send(ctx, result)
			return
		}

		logger.Info("Step 2: Checking inventory", zap.String("orderID", input.OrderID))
		var inventoryResult InventoryCheckResult
		err = workflow.ExecuteActivity(ctx, CheckInventoryActivity, input).Get(ctx, &inventoryResult)
		if err != nil {
			logger.Error("Failed to check inventory", zap.Error(err), zap.String("orderID", input.OrderID))
			result = &OrderWorkflowOutput{
				OrderID: input.OrderID,
				Status:  order.OrderStatusFailed,
				Error:   fmt.Sprintf("Failed to check inventory: %v", err),
			}
			resultCh.Send(ctx, result)
			return
		}

		if !inventoryResult.Available {
			logger.Error("Inventory check failed", 
				zap.String("orderID", input.OrderID),
				zap.String("message", inventoryResult.Message),
			)
			result = &OrderWorkflowOutput{
				OrderID: input.OrderID,
				Status:  order.OrderStatusFailed,
				Error:   inventoryResult.Message,
			}
			resultCh.Send(ctx, result)
			return
		}

		logger.Info("Step 3: Processing payment", zap.String("orderID", input.OrderID))
		var paymentResult PaymentResult
		err = workflow.ExecuteActivity(ctx, ProcessPaymentActivity, input).Get(ctx, &paymentResult)
		if err != nil {
			logger.Error("Failed to process payment", zap.Error(err), zap.String("orderID", input.OrderID))
			result = &OrderWorkflowOutput{
				OrderID: input.OrderID,
				Status:  order.OrderStatusFailed,
				Error:   fmt.Sprintf("Failed to process payment: %v", err),
			}
			resultCh.Send(ctx, result)
			return
		}

		if paymentResult.Status != "completed" {
			logger.Error("Payment processing failed", 
				zap.String("orderID", input.OrderID),
				zap.String("status", paymentResult.Status),
				zap.String("message", paymentResult.Message),
			)
			result = &OrderWorkflowOutput{
				OrderID: input.OrderID,
				Status:  order.OrderStatusFailed,
				Error:   paymentResult.Message,
			}
			resultCh.Send(ctx, result)
			return
		}

		logger.Info("Step 4: Sending notification", zap.String("orderID", input.OrderID))
		var notificationResult NotificationResult
		err = workflow.ExecuteActivity(ctx, SendNotificationActivity, input, paymentResult).Get(ctx, &notificationResult)
		if err != nil {
			logger.Error("Failed to send notification", zap.Error(err), zap.String("orderID", input.OrderID))
			logger.Warn("Notification failed, but order processing completed successfully", 
				zap.Error(err),
				zap.String("orderID", input.OrderID),
			)
		}

		logger.Info("Order workflow completed successfully", 
			zap.String("orderID", input.OrderID),
			zap.String("status", string(order.OrderStatusCompleted)),
		)

		result = &OrderWorkflowOutput{
			OrderID:       input.OrderID,
			Status:        order.OrderStatusCompleted,
			PaymentID:     paymentResult.PaymentID,
			TransactionID: paymentResult.TransactionID,
		}
		resultCh.Send(ctx, result)
	})

	selector.AddReceive(cancelSignal, func(c workflow.ReceiveChannel, more bool) {
		logger.Info("Received cancellation signal", zap.String("orderID", input.OrderID))
		resultCh.Send(ctx, &OrderWorkflowOutput{
			OrderID: input.OrderID,
			Status:  order.OrderStatusCancelled,
			Error:   "Order was cancelled by user",
		})
	})

	selector.AddReceive(resultCh, func(c workflow.ReceiveChannel, more bool) {
	})

	selector.Select(ctx)

	var result OrderWorkflowOutput
	resultCh.Receive(ctx, &result)

	return &result, nil
}
