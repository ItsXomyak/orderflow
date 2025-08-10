package activity

import (
	"context"
	"time"

	"go.temporal.io/sdk/activity"

	"orderflow/internal/domain/inventory"
	"orderflow/internal/domain/order"
	"orderflow/internal/domain/workflow"
)

// чек наличия товаров на складе
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

// чек наличия товаров
func (a *CheckInventoryActivity) Execute(ctx context.Context, input *workflow.CheckInventoryActivityInput) (*workflow.CheckInventoryActivityOutput, error) {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting CheckInventoryActivity", "order_id", input.OrderID)

	// валидация входных данных
	if err := input.Validate(); err != nil {
		logger.Error("Validation failed", "error", err)
		return nil, workflow.NewActivityError(
			workflow.CheckInventoryActivity,
			workflow.StepCheckInventory,
			workflow.ErrorCodeValidation,
			err.Error(),
			false,
		)
	}

	// апдейт статус заказа
	if err := a.orderService.UpdateStatus(ctx, input.OrderID, order.StatusValidating); err != nil {
		logger.Error("Failed to update order status", "error", err)
		// продолжаем выполнение, это не критичная ошибка
	}

	// симуляция времени проверки склада
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(workflow.InventoryCheckDuration):
	}

	// подготовка запроса для проверки склада
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

	// проверка наличия товаров
	checkResp, err := a.inventoryService.CheckAvailability(ctx, checkReq)
	if err != nil {
		logger.Error("Failed to check inventory", "error", err)
		
		// обновляем статус заказа при ошибке
		a.orderService.SetFailure(ctx, input.OrderID, "Failed to check inventory: "+err.Error())
		
		// можно ли повторить операцию?
		retryable := true
		errorCode := workflow.ErrorCodeInternalError
		
		switch err.(type) {
		case *inventory.ProductNotFoundError:
			retryable = false
		}
		
		return nil, workflow.NewActivityError(
			workflow.CheckInventoryActivity,
			workflow.StepCheckInventory,
			errorCode,
			err.Error(),
			retryable,
		)
	}

	// если товары недоступны
	if !checkResp.Available {
		logger.Warn("Inventory not available", "unavailable_items", checkResp.UnavailableItems)
		
		// апдейт статус заказа
		a.orderService.SetFailure(ctx, input.OrderID, "Some items are not available")
		
		return &workflow.CheckInventoryActivityOutput{
			Available:        false,
			UnavailableItems: checkResp.UnavailableItems,
		}, nil
	}

	// резерв товары
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
		
		return nil, workflow.NewActivityError(
			workflow.CheckInventoryActivity,
			workflow.StepCheckInventory,
			workflow.ErrorCodeInventoryUnavailable,
			err.Error(),
			true,
		)
	}

	logger.Info("Inventory checked and items reserved successfully", "order_id", input.OrderID)

	return &workflow.CheckInventoryActivityOutput{
		Available:        true,
		UnavailableItems: nil,
	}, nil
}

// имя Activity для регистрации в Temporal
func (a *CheckInventoryActivity) GetActivityName() string {
	return workflow.CheckInventoryActivity
}