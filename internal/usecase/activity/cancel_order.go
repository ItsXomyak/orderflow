// internal/usecase/activity/cancel_order.go
package activity

import (
	"context"

	"go.temporal.io/sdk/activity"

	"orderflow/internal/domain/inventory"
	"orderflow/internal/domain/order"
	"orderflow/internal/domain/payment"
	"orderflow/internal/domain/workflow"
)

type CancelOrderActivityInput struct {
	OrderID    string `json:"order_id"`
	CustomerID string `json:"customer_id"`
	Reason     string `json:"reason"`
}

type CancelOrderActivity struct {
	orderService     order.Service
	paymentService   payment.Service
	inventoryService inventory.Service
}

func NewCancelOrderActivity(
	orderService order.Service,
	paymentService payment.Service,
	inventoryService inventory.Service,
) *CancelOrderActivity {
	return &CancelOrderActivity{
		orderService:     orderService,
		paymentService:   paymentService,
		inventoryService: inventoryService,
	}
}

// фукнция отмена заказа
func (a *CancelOrderActivity) Execute(ctx context.Context, input *CancelOrderActivityInput) error {
	logger := activity.GetLogger(ctx)
	logger.Info("Starting CancelOrderActivity", 
		"order_id", input.OrderID,
		"customer_id", input.CustomerID,
		"reason", input.Reason)

	// валидация входных данных
	if input.OrderID == "" {
		return workflow.NewActivityError(
			workflow.CancelOrderActivity,
			workflow.StepCancelled,
			workflow.ErrorCodeValidation,
			"order_id is required",
			false,
		)
	}

	// получить заказ
	orderEntity, err := a.orderService.GetByID(ctx, input.OrderID)
	if err != nil {
		logger.Error("Failed to get order", "error", err)
		
		errorCode := workflow.ErrorCodeInternalError
		if _, ok := err.(*order.NotFoundError); ok {
			errorCode = workflow.ErrorCodeOrderNotFound
		}
		
		return workflow.NewActivityError(
			workflow.CancelOrderActivity,
			workflow.StepCancelled,
			errorCode,
			err.Error(),
			false,
		)
	}

	// чек, можно ли отменить заказ
	if !orderEntity.CanBeCancelled() {
		logger.Warn("Cannot cancel order", "order_status", orderEntity.Status)
		return workflow.NewActivityError(
			workflow.CancelOrderActivity,
			workflow.StepCancelled,
			workflow.ErrorCodeOrderCancelled,
			"Order cannot be cancelled in current status: "+string(orderEntity.Status),
			false,
		)
	}

	// шаги отмены заказа (в обратном порядке выполнения)
	
	// возврат средств (если платеж был произведен)
	if orderEntity.PaymentID != "" {
		logger.Info("Refunding payment", "payment_id", orderEntity.PaymentID)
		
		refundReq := &payment.RefundRequest{
			PaymentID: orderEntity.PaymentID,
			Reason:    "Order cancelled: " + input.Reason,
		}
		
		if err := a.paymentService.RefundPayment(ctx, refundReq); err != nil {
			logger.Error("Failed to refund payment", "error", err, "payment_id", orderEntity.PaymentID)
			// не останавливаем процесс отмены из-за ошибки возврата - логируем для ручной обработки
			// в прода нужно реализовать очередь для повторных попыток возврата
		}
	}

	// освобождение резервирования товаров
	logger.Info("Releasing inventory reservation", "order_id", input.OrderID)
	
	if err := a.inventoryService.ReleaseReservation(ctx, input.OrderID); err != nil {
		logger.Error("Failed to release inventory reservation", "error", err)
	}

	// апдейт статуса заказа
	logger.Info("Updating order status to cancelled", "order_id", input.OrderID)
	
	if err := a.orderService.Cancel(ctx, input.OrderID); err != nil {
		logger.Error("Failed to cancel order", "error", err)
		return workflow.NewActivityError(
			workflow.CancelOrderActivity,
			workflow.StepCancelled,
			workflow.ErrorCodeInternalError,
			"Failed to update order status: "+err.Error(),
			true,
		)
	}

	logger.Info("Order cancelled successfully", "order_id", input.OrderID)
	return nil
}

func (a *CancelOrderActivity) GetActivityName() string {
	return workflow.CancelOrderActivity
}