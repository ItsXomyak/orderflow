package service

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"orderflow/internal/domain/payment"
	"orderflow/pkg/logger"
)

type PaymentService struct {
	paymentRepo payment.Repository
}

func NewPaymentService(paymentRepo payment.Repository) *PaymentService {
	return &PaymentService{
		paymentRepo: paymentRepo,
	}
}

func (service *PaymentService) ProcessPayment(ctx context.Context, req *payment.Request) (*payment.Response, error) {
	logger.Info("Processing payment", "order_id", req.OrderID, "amount", req.Amount, "currency", req.Currency)

	if req.OrderID == "" {
		return nil, payment.NewValidationError("order_id is required")
	}
	if req.CustomerID == "" {
		return nil, payment.NewValidationError("customer_id is required")
	}
	if req.Amount <= 0 {
		return nil, payment.NewValidationError("amount must be positive")
	}
	if req.Currency == "" {
		return nil, payment.NewValidationError("currency is required")
	}

	existingPayment, err := service.paymentRepo.GetPaymentByOrderID(ctx, req.OrderID)
	if err == nil && existingPayment != nil {
		return nil, payment.NewDuplicatePaymentError(req.OrderID)
	}

	paymentEntity := payment.NewPayment(req)
	paymentEntity.ID = uuid.New().String()

	success, transactionID, errorCode, errorMessage := service.simulatePaymentProcessing(req)

	if success {
		paymentEntity.Complete(transactionID)
		logger.Info("Payment processed successfully", 
			"payment_id", paymentEntity.ID, 
			"transaction_id", transactionID,
			"order_id", req.OrderID)
	} else {
		paymentEntity.Fail(errorMessage)
		logger.Error("Payment failed", 
			"payment_id", paymentEntity.ID,
			"error_code", errorCode,
			"error_message", errorMessage,
			"order_id", req.OrderID)
	}

	if err := service.paymentRepo.CreatePayment(ctx, paymentEntity); err != nil {
		return nil, err
	}

	response := &payment.Response{
		Success:       success,
		PaymentID:     paymentEntity.ID,
		TransactionID: transactionID,
		ErrorCode:     errorCode,
		ErrorMessage:  errorMessage,
	}

	return response, nil
}

func (service *PaymentService) simulatePaymentProcessing(req *payment.Request) (bool, string, string, string) {
	
	if time.Now().UnixNano()%10 < 9 {
		transactionID := fmt.Sprintf("txn_%s", uuid.New().String()[:8])
		return true, transactionID, "", ""
	}
	
	failureScenarios := []struct {
		code    string
		message string
	}{
		{"INSUFFICIENT_FUNDS", "Insufficient funds on the card"},
		{"CARD_DECLINED", "Card was declined by the bank"},
		{"EXPIRED_CARD", "Card has expired"},
		{"INVALID_CVV", "Invalid CVV code"},
		{"NETWORK_ERROR", "Payment network temporarily unavailable"},
	}
	
	scenario := failureScenarios[time.Now().UnixNano()%int64(len(failureScenarios))]
	return false, "", scenario.code, scenario.message
}

func (service *PaymentService) GetPayment(ctx context.Context, paymentID string) (*payment.Payment, error) {
	if paymentID == "" {
		return nil, payment.NewValidationError("payment_id is required")
	}

	return service.paymentRepo.GetPayment(ctx, paymentID)
}

func (service *PaymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	if orderID == "" {
		return nil, payment.NewValidationError("order_id is required")
	}

	return service.paymentRepo.GetPaymentByOrderID(ctx, orderID)
}

func (service *PaymentService) RefundPayment(ctx context.Context, req *payment.RefundRequest) error {
	logger.Info("Processing refund", "payment_id", req.PaymentID, "amount", req.Amount, "reason", req.Reason)

	if req.PaymentID == "" {
		return payment.NewValidationError("payment_id is required")
	}

	paymentEntity, err := service.paymentRepo.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return err
	}

	if err := paymentEntity.Refund(); err != nil {
		return err
	}

	success := service.simulateRefundProcessing(paymentEntity)
	if !success {
		logger.Error("Refund failed", "payment_id", req.PaymentID)
		return payment.NewRefundFailedError(req.PaymentID, "Refund processing failed")
	}

	if err := service.paymentRepo.UpdatePayment(ctx, paymentEntity); err != nil {
		return err
	}

	logger.Info("Refund processed successfully", "payment_id", req.PaymentID)
	return nil
}

func (service *PaymentService) simulateRefundProcessing(paymentEntity *payment.Payment) bool {
	return time.Now().UnixNano()%20 < 19
}

func (service *PaymentService) CancelPayment(ctx context.Context, paymentID string) error {
	logger.Info("Cancelling payment", "payment_id", paymentID)

	if paymentID == "" {
		return payment.NewValidationError("payment_id is required")
	}

	paymentEntity, err := service.paymentRepo.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}

	if paymentEntity.Status != payment.StatusPending {
		return payment.NewCannotCancelError(paymentID, paymentEntity.Status)
	}

	paymentEntity.Fail("Payment cancelled by user")
	
	if err := service.paymentRepo.UpdatePayment(ctx, paymentEntity); err != nil {
		return err
	}

	logger.Info("Payment cancelled", "payment_id", paymentID)
	return nil
}

func (service *PaymentService) GetPayments(ctx context.Context) ([]*payment.Payment, error) {
	return service.paymentRepo.GetPayments(ctx)
}

func (service *PaymentService) GetPaymentStatistics(ctx context.Context) (*PaymentStatistics, error) {
	payments, err := service.paymentRepo.GetPayments(ctx)
	if err != nil {
		return nil, err
	}

	stats := &PaymentStatistics{
		TotalPayments:    0,
		CompletedPayments: 0,
		FailedPayments:    0,
		RefundedPayments:  0,
		TotalAmount:       0,
		RefundedAmount:    0,
	}

	for _, paymentEntity := range payments {
		stats.TotalPayments++
		stats.TotalAmount += paymentEntity.Amount

		switch paymentEntity.Status {
		case payment.StatusCompleted:
			stats.CompletedPayments++
		case payment.StatusFailed:
			stats.FailedPayments++
		case payment.StatusRefunded:
			stats.RefundedPayments++
			stats.RefundedAmount += paymentEntity.Amount
		}
	}

	if stats.TotalPayments > 0 {
		stats.SuccessRate = float64(stats.CompletedPayments) / float64(stats.TotalPayments) * 100
	}

	return stats, nil
}

type PaymentStatistics struct {
	TotalPayments     int     `json:"total_payments"`
	CompletedPayments int     `json:"completed_payments"`
	FailedPayments    int     `json:"failed_payments"`
	RefundedPayments  int     `json:"refunded_payments"`
	TotalAmount       float64 `json:"total_amount"`
	RefundedAmount    float64 `json:"refunded_amount"`
	SuccessRate       float64 `json:"success_rate"`
}
