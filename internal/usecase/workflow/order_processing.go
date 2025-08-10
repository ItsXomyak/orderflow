package workflow

import (
	"time"

	"go.temporal.io/sdk/temporal"
	"go.temporal.io/sdk/workflow"

	"orderflow/internal/domain/notification"
	"orderflow/internal/domain/order"
	workflowDomain "orderflow/internal/domain/workflow"
	"orderflow/internal/usecase/activity"
)

func OrderProcessingWorkflow(ctx workflow.Context, input *workflowDomain.OrderProcessingInput) (*workflowDomain.WorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Starting OrderProcessingWorkflow", "customer_id", input.CustomerID)

	state := workflowDomain.NewState("", input.CustomerID)

	retryPolicy := &temporal.RetryPolicy{
		InitialInterval:    workflowDomain.DefaultInitialInterval,
		BackoffCoefficient: workflowDomain.DefaultBackoffCoefficient,
		MaximumInterval:    workflowDomain.DefaultMaximumInterval,
		MaximumAttempts:    workflowDomain.DefaultMaximumAttempts,
	}

	activityOptions := workflow.ActivityOptions{
		StartToCloseTimeout: 30 * time.Second,
		RetryPolicy:         retryPolicy,
	}

	ctx = workflow.WithActivityOptions(ctx, activityOptions)

	cancelChannel := workflow.GetSignalChannel(ctx, workflowDomain.CancelOrderSignal)

	err := workflow.SetQueryHandler(ctx, workflowDomain.OrderStatusQuery, func() (order.Status, error) {
		return state.Status, nil
	})
	if err != nil {
		logger.Error("Failed to set order status query handler", "error", err)
		return nil, err
	}

	err = workflow.SetQueryHandler(ctx, workflowDomain.WorkflowStateQuery, func() (*workflowDomain.State, error) {
		return state, nil
	})
	if err != nil {
		logger.Error("Failed to set workflow state query handler", "error", err)
		return nil, err
	}

	var orderID string
	var paymentID string

	logger.Info("Step 1: Creating order")
	state.UpdateStep(workflowDomain.StepCreateOrder)

	createOrderInput := &workflowDomain.CreateOrderActivityInput{
		CustomerID: input.CustomerID,
		Items:      input.Items,
	}

	var createOrderOutput *workflowDomain.CreateOrderActivityOutput

	selector := workflow.NewSelector(ctx)
	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
		var signal string
		c.Receive(ctx, &signal)
		logger.Info("Received cancel signal", "signal", signal)
		state.Cancel()
	})

	createOrderFuture := workflow.ExecuteActivity(ctx, workflowDomain.CreateOrderActivity, createOrderInput)
	selector.AddFuture(createOrderFuture, func(f workflow.Future) {
		if err := f.Get(ctx, &createOrderOutput); err != nil {
			logger.Error("Create order failed", "error", err)
			state.SetError(workflowDomain.ErrorCodeInternalError, err.Error())
		}
	})

	selector.Select(ctx)

	if state.IsCancelled {
		return handleCancellation(ctx, "", "")
	}

	if state.IsFailed() {
		return handleFailure(ctx, state, "", input.CustomerID)
	}

	orderID = createOrderOutput.OrderID
	state.OrderID = orderID
	logger.Info("Order created successfully", "order_id", orderID)

	logger.Info("Step 2: Checking inventory")
	state.UpdateStep(workflowDomain.StepCheckInventory)

	checkInventoryInput := &workflowDomain.CheckInventoryActivityInput{
		OrderID: orderID,
		Items:   input.Items,
	}

	var checkInventoryOutput *workflowDomain.CheckInventoryActivityOutput

	selector = workflow.NewSelector(ctx)
	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
		var signal string
		c.Receive(ctx, &signal)
		logger.Info("Received cancel signal", "signal", signal)
		state.Cancel()
	})

	checkInventoryFuture := workflow.ExecuteActivity(ctx, workflowDomain.CheckInventoryActivity, checkInventoryInput)
	selector.AddFuture(checkInventoryFuture, func(f workflow.Future) {
		if err := f.Get(ctx, &checkInventoryOutput); err != nil {
			logger.Error("Check inventory failed", "error", err)
			state.SetError(workflowDomain.ErrorCodeInventoryUnavailable, err.Error())
		}
	})

	selector.Select(ctx)

	if state.IsCancelled {
		return handleCancellation(ctx, orderID, input.CustomerID)
	}

	if state.IsFailed() {
		return handleFailure(ctx, state, orderID, input.CustomerID)
	}

	if !checkInventoryOutput.Available {
		logger.Warn("Inventory not available", "unavailable_items", checkInventoryOutput.UnavailableItems)
		state.SetError(workflowDomain.ErrorCodeInventoryUnavailable, "Some items are not available")
		return handleFailure(ctx, state, orderID, input.CustomerID)
	}

	logger.Info("Inventory check passed", "order_id", orderID)

	logger.Info("Step 3: Processing payment")
	state.UpdateStep(workflowDomain.StepProcessPayment)

	var totalAmount float64
	for _, item := range input.Items {
		totalAmount += item.Price * float64(item.Quantity)
	}

	processPaymentInput := &workflowDomain.ProcessPaymentActivityInput{
		OrderID:    orderID,
		CustomerID: input.CustomerID,
		Amount:     totalAmount,
		Currency:   "USD",
	}

	var processPaymentOutput *workflowDomain.ProcessPaymentActivityOutput

	// проверяем сигнал отмены
	selector = workflow.NewSelector(ctx)
	selector.AddReceive(cancelChannel, func(c workflow.ReceiveChannel, more bool) {
		var signal string
		c.Receive(ctx, &signal)
		logger.Info("Received cancel signal", "signal", signal)
		state.Cancel()
	})

	processPaymentFuture := workflow.ExecuteActivity(ctx, workflowDomain.ProcessPaymentActivity, processPaymentInput)
	selector.AddFuture(processPaymentFuture, func(f workflow.Future) {
		if err := f.Get(ctx, &processPaymentOutput); err != nil {
			logger.Error("Process payment failed", "error", err)
			state.SetError(workflowDomain.ErrorCodePaymentFailed, err.Error())
		}
	})

	selector.Select(ctx)

	if state.IsCancelled {
		return handleCancellation(ctx, orderID, input.CustomerID)
	}

	if state.IsFailed() {
		return handleFailure(ctx, state, orderID, input.CustomerID)
	}

	paymentID = processPaymentOutput.PaymentID
	state.PaymentID = paymentID
	logger.Info("Payment processed successfully", "order_id", orderID, "payment_id", paymentID)

	logger.Info("Step 4: Sending notification")
	state.UpdateStep(workflowDomain.StepSendNotification)

	sendNotificationInput := &workflowDomain.SendNotificationActivityInput{
		CustomerID: input.CustomerID,
		OrderID:    orderID,
		Type:       notification.TypeOrderConfirmed,
		Channel:    notification.ChannelEmail,
		Message:    "",
	}

	err = workflow.ExecuteActivity(ctx, workflowDomain.SendNotificationActivity, sendNotificationInput).Get(ctx, nil)
	if err != nil {
		logger.Warn("Failed to send notification, but continuing workflow", "error", err)
	} else {
		logger.Info("Notification sent successfully", "order_id", orderID)
	}

	logger.Info("Step 5: Completing workflow")
	state.UpdateStep(workflowDomain.StepComplete)
	state.UpdateStatus(order.StatusCompleted)

	logger.Info("OrderProcessingWorkflow completed successfully",
		"order_id", orderID,
		"payment_id", paymentID,
		"duration", state.GetDuration())

	return &workflowDomain.WorkflowResult{
		OrderID:   orderID,
		Status:    order.StatusCompleted,
		Success:   true,
		PaymentID: paymentID,
		Message:   "Order processed successfully",
	}, nil
}

func handleCancellation(ctx workflow.Context, orderID, customerID string) (*workflowDomain.WorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Info("Handling order cancellation", "order_id", orderID)

	if orderID != "" {
		cancelInput := &activity.CancelOrderActivityInput{
			OrderID:    orderID,
			CustomerID: customerID,
			Reason:     "Customer cancellation",
		}

		err := workflow.ExecuteActivity(ctx, workflowDomain.CancelOrderActivity, cancelInput).Get(ctx, nil)
		if err != nil {
			logger.Error("Failed to cancel order", "error", err, "order_id", orderID)
		}

		notificationInput := &workflowDomain.SendNotificationActivityInput{
			CustomerID: customerID,
			OrderID:    orderID,
			Type:       notification.TypeOrderCancelled,
			Channel:    notification.ChannelEmail,
		}

		err = workflow.ExecuteActivity(ctx, workflowDomain.SendNotificationActivity, notificationInput).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send cancellation notification", "error", err)
		}
	}

	return &workflowDomain.WorkflowResult{
		OrderID: orderID,
		Status:  order.StatusCancelled,
		Success: false,
		Message: "Order was cancelled",
	}, nil
}

func handleFailure(
	ctx workflow.Context,
	state *workflowDomain.State,
	orderID,
	customerID string,
) (*workflowDomain.WorkflowResult, error) {
	logger := workflow.GetLogger(ctx)
	logger.Error("Handling order failure",
		"order_id", orderID,
		"error_code", state.ErrorCode,
		"error_message", state.ErrorMessage)

	if orderID != "" && customerID != "" {
		notificationInput := &workflowDomain.SendNotificationActivityInput{
			CustomerID: customerID,
			OrderID:    orderID,
			Type:       notification.TypeOrderFailed,
			Channel:    notification.ChannelEmail,
			Message:    state.ErrorMessage,
		}

		err := workflow.ExecuteActivity(ctx, workflowDomain.SendNotificationActivity, notificationInput).Get(ctx, nil)
		if err != nil {
			logger.Warn("Failed to send failure notification", "error", err)
		}
	}

	return &workflowDomain.WorkflowResult{
		OrderID: orderID,
		Status:  order.StatusFailed,
		Success: false,
		Message: state.ErrorMessage,
	}, workflowDomain.NewActivityError("OrderProcessingWorkflow", state.CurrentStep, state.ErrorCode, state.ErrorMessage, false)
}
