package activity

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"

	"orderflow/internal/domain/inventory"
	"orderflow/internal/domain/order"
	wf "orderflow/internal/domain/workflow"
)

type CheckInventoryActivity struct {
	inventoryService inventory.Service
	orderService     order.Service
}

func NewCheckInventoryActivity(inventoryService inventory.Service, orderService order.Service) *CheckInventoryActivity {
	return &CheckInventoryActivity{
		inventoryService: inventoryService,
		orderService:     orderService,
	}
}

func (a *CheckInventoryActivity) Execute(ctx context.Context, input *wf.CheckInventoryActivityInput) (*wf.CheckInventoryActivityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting CheckInventoryActivity", "order_id", input.OrderID)

	if err := input.Validate(); err != nil {
		logger.Error("Validation failed", "error", err)
		return nil, wf.NewActivityError(
			wf.CheckInventoryActivity,
			wf.StepCheckInventory,
			wf.ErrorCodeValidation,
			err.Error(),
			false,
		)
	}

	if err := a.orderService.UpdateStatus(ctx, input.OrderID, order.StatusValidating); err != nil {
		logger.Error("Failed to update order status", "error", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(wf.InventoryCheckDuration):
	}

	checkItems := make([]inventory.CheckItem, len(input.Items))
	for i, item := range input.Items {
		checkItems[i] = inventory.CheckItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	checkReq := &inventory.CheckRequest{
		OrderID: input.OrderID,
		Items:   checkItems,
	}

	checkResp, err := a.inventoryService.CheckAvailability(ctx, checkReq)
	if err != nil {
		logger.Error("Failed to check inventory", "error", err)
		
		a.orderService.SetFailure(ctx, input.OrderID, "Failed to check inventory: "+err.Error())
		
		retryable := true
		errorCode := wf.ErrorCodeInternalError
		
		switch err.(type) {
		case *inventory.ProductNotFoundError:
			retryable = false
		}
		
		return nil, wf.NewActivityError(
			wf.CheckInventoryActivity,
			wf.StepCheckInventory,
			errorCode,
			err.Error(),
			retryable,
		)
	}

	if !checkResp.Available {
		logger.Warn("Inventory not available", "unavailable_items", checkResp.UnavailableItems)
		
		a.orderService.SetFailure(ctx, input.OrderID, "Some items are not available")
		
		return &wf.CheckInventoryActivityOutput{
			Available:        false,
			UnavailableItems: checkResp.UnavailableItems,
		}, nil
	}

	reserveItems := make([]inventory.ReserveItem, len(input.Items))
	for i, item := range input.Items {
		reserveItems[i] = inventory.ReserveItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		}
	}

	reserveReq := &inventory.ReserveRequest{
		OrderID: input.OrderID,
		Items:   reserveItems,
	}

	if err := a.inventoryService.ReserveItems(ctx, reserveReq); err != nil {
		logger.Error("Failed to reserve items", "error", err)
		
		a.orderService.SetFailure(ctx, input.OrderID, "Failed to reserve items: "+err.Error())
		
		return nil, wf.NewActivityError(
			wf.CheckInventoryActivity,
			wf.StepCheckInventory,
			wf.ErrorCodeInventoryUnavailable,
			err.Error(),
			true,
		)
	}

	logger.Info("Inventory checked and items reserved successfully", "order_id", input.OrderID)

	return &wf.CheckInventoryActivityOutput{
		Available:        true,
		UnavailableItems: nil,
	}, nil
}

func (a *CheckInventoryActivity) GetActivityName() (string, error) {
	return wf.CheckInventoryActivity, nil
}