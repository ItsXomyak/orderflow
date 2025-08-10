package activity

import (
	"context"
	"time"

	"orderflow/internal/domain/order"
	"orderflow/internal/domain/workflow"
	"orderflow/pkg/logger"
)

type CreateOrderActivity struct {
	orderService order.Service
}

func NewCreateOrderActivity(orderService order.Service) *CreateOrderActivity {
	return &CreateOrderActivity{orderService: orderService}
}

func (a *CreateOrderActivity) Execute(ctx context.Context, input *workflow.CreateOrderActivityInput) (*workflow.CreateOrderActivityOutput, error) {
	logger.Info("Starting CreateOrderActivity", "customer_id", input.CustomerID)

	if err := input.Validate(); err != nil {
		logger.Error("CreateOrderActivity validation error", "error", err)
		return nil, workflow.NewActivityError(
			workflow.CreateOrderActivity,
			workflow.StepCreateOrder,
			workflow.ErrorCodeValidation,
			err.Error(),
			false, // не ретраить в таком случае
		)
	}


	// симуляция процесса создания заказа
	select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(workflow.OrderCreationDuration):
		}


		// запрос на заказ
		createReq := &order.CreateRequest{
			CustomerID: input.CustomerID,
			Items:      input.Items,
		}

		// создаем заказ через сервис
		newOrder, err := a.orderService.Create(ctx, createReq)
		if err != nil {
			logger.Error("Failed to create order", "error", err)

			// чекаем можно ли повторить операцию
			retryable := true
			errorCode:= workflow.ErrorCodeInternalError

			switch err.(type) {
			case *order.ValidationError:
				retryable = false
				errorCode = workflow.ErrorCodeInternalError
			}

			return nil, workflow.NewActivityError(
				workflow.CreateOrderActivity,
				workflow.StepCreateOrder,
				errorCode,
				err.Error(),
				retryable,
			)
		}

		logger.Info("Order created successfully", "order_id", newOrder.ID)

		return &workflow.CreateOrderActivityOutput{
			OrderID: newOrder.ID,
		}, nil	
}

func (a *CreateOrderActivity) GetActivityName() string {
	return workflow.CreateOrderActivity
}