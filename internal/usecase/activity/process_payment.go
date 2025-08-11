package activity

import (
	"context"
	"time"

	"orderflow/internal/domain/inventory"
	"orderflow/internal/domain/order"
	"orderflow/internal/domain/payment"
	wf "orderflow/internal/domain/workflow"
	"orderflow/pkg/logger"
)

type ProcessPaymentActivity struct {
	paymenyService   payment.Service
	orderService     order.Service
	inventoryService inventory.Service
}

func NewProcessPaymentActivity(paymenyService payment.Service, orderService order.Service, inventoryService inventory.Service) *ProcessPaymentActivity {
	return &ProcessPaymentActivity{paymenyService: paymenyService, orderService: orderService, inventoryService: inventoryService}
}

func (a *ProcessPaymentActivity) Execute(ctx context.Context, input *wf.ProcessPaymentActivityInput) (*wf.ProcessPaymentActivityOutput, error) {
	logger.Info("Starting ProcessPaymentActivity", "order_id", input.OrderID)

	if err := input.Validate(); err != nil {
		logger.Error("ProcessPaymentActivity validation error", "error", err)
		return nil, wf.NewActivityError(
			wf.ProcessPaymentActivity,
			wf.StepProcessPayment,
			wf.ErrorCodeValidation,
			err.Error(),
			false, // не ретраить в таком случае
		)
	}

	if err := a.orderService.UpdateStatus(ctx, input.OrderID, order.StatusPayment); err != nil {
		logger.Error("Failed to update order status", "error", err)
	}

	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(wf.PaymentProcessDuration):
	}

	paymentReq := &payment.Request{
		OrderID:   input.OrderID,
		CustomerID: input.CustomerID,
		Amount:    input.Amount,
		Currency:  input.Currency,
		PaymentMethod: "card", // если нужно то расширить, пока дефолт
	}

	paymentResp, err := a.paymenyService.ProcessPayment(ctx, paymentReq)
	if err != nil {
		logger.Error("Failed to process payment", "error", err)
		if releaseErr := a.inventoryService.ReleaseReservation(ctx, input.OrderID); releaseErr != nil {
			logger.Error("Failed to release reservation after payment failure", "error", releaseErr)
		}

		a.orderService.SetFailure(ctx, input.OrderID, "Payment failed"+ err.Error())

		retryable := true
		errorCode := wf.ErrorCodePaymentFailed

		switch err.(type) {
		case *payment.ValidationError:
			errorCode = wf.ErrorCodeValidation
			retryable = false
		case *payment.InsufficientFundsError: 
		retryable = false
		}

		return nil, wf.NewActivityError(
			wf.ProcessPaymentActivity,
			wf.StepProcessPayment,
			errorCode,
			err.Error(),
			retryable,
		)
	}

	if !paymentResp.Success {
		logger.Error("Payment was not successful", "error_code", paymentResp.ErrorCode, "error_message", paymentResp.ErrorMessage)
		
		if releaseErr := a.inventoryService.ReleaseReservation(ctx, input.OrderID); releaseErr != nil {
			logger.Error("Failed to release reservation after payment failure", "error", releaseErr)
		}
		
		a.orderService.SetFailure(ctx, input.OrderID, paymentResp.ErrorMessage)
		
		return nil, wf.NewActivityError(
			wf.ProcessPaymentActivity,
			wf.StepProcessPayment,
			wf.ErrorCodePaymentFailed,
			paymentResp.ErrorMessage,
			false, // обычно ошибки платежа не ретраим автоматически
		)
	}

	if err := a.inventoryService.ConfirmReservation(ctx, input.OrderID); err != nil {
		logger.Error("Failed to confirm reservation", "error", err)
		
		return nil, wf.NewActivityError(
			wf.ProcessPaymentActivity,
			wf.StepProcessPayment,
			wf.ErrorCodeInternalError,
			"Failed to confirm reservation after successful payment: "+err.Error(),
			true,
		)
	}

	logger.Info("Payment processed successfully", 
		"order_id", input.OrderID, 
		"payment_id", paymentResp.PaymentID,
		"transaction_id", paymentResp.TransactionID)

	return &wf.ProcessPaymentActivityOutput{
		PaymentID:     paymentResp.PaymentID,
		TransactionID: paymentResp.TransactionID,
	}, nil
}


func (a *ProcessPaymentActivity) GetActivityName() (string, error) {
	return wf.ProcessPaymentActivity, nil
}
