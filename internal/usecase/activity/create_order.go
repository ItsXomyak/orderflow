package activity

import (
	"context"
	"time"

	"go.temporal.io/sdk/temporal"

	"orderflow/internal/domain/order"
	wf "orderflow/internal/domain/workflow"
	"orderflow/pkg/logger"
)

type CreateOrderActivity struct {
	orderService order.Service
}

func NewCreateOrderActivity(svc order.Service) *CreateOrderActivity {
	return &CreateOrderActivity{orderService: svc}
}

func (a *CreateOrderActivity) Execute(ctx context.Context, in *wf.CreateOrderActivityInput) (*wf.CreateOrderActivityOutput, error) {
	if in == nil {
		return nil, temporal.NewNonRetryableApplicationError("nil input", wf.ErrorCodeValidation, nil)
	}

	logger.Info("CreateOrderActivity: start", "customer_id", in.CustomerID)
	if v, ok := any(in).(interface{ Validate() error }); ok {
		if err := v.Validate(); err != nil {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), wf.ErrorCodeValidation, nil)
		}
	} else {
		if in.CustomerID == "" || len(in.Items) == 0 {
			return nil, temporal.NewNonRetryableApplicationError("customer_id and items are required", wf.ErrorCodeValidation, nil)
		}
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(wf.OrderCreationDuration):
	}

	req := &order.CreateRequest{
		CustomerID: in.CustomerID,
		Items:      in.Items, // <- должен быть []order.Item
	}

	o, err := a.orderService.Create(ctx, req)
	if err != nil {
		if _, ok := err.(*order.ValidationError); ok {
			return nil, temporal.NewNonRetryableApplicationError(err.Error(), wf.ErrorCodeValidation, nil)
		}
		return nil, temporal.NewApplicationError(err.Error(), wf.ErrorCodeInternalError)
	}

	logger.Info("CreateOrderActivity: success", "order_id", o.ID)
	return &wf.CreateOrderActivityOutput{OrderID: o.ID}, nil
}

func (a *CreateOrderActivity) GetActivityName() string {
	return wf.CreateOrderActivity
}
