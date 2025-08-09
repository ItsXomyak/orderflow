package payment

import (
	"fmt"
	"time"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusProcessing PaymentStatus = "processing"
	PaymentStatusCompleted PaymentStatus = "completed"
	PaymentStatusFailed    PaymentStatus = "failed"
)

type PaymentMethod string

const (
	PaymentMethodCreditCard PaymentMethod = "credit_card"
	PaymentMethodPayPal     PaymentMethod = "paypal"
	PaymentMethodBankTransfer PaymentMethod = "bank_transfer"
)

type Payment struct {
	ID            string        `json:"id"`
	OrderID       string        `json:"order_id"`
	CustomerID    string        `json:"customer_id"`
	Amount        float64       `json:"amount"`
	Currency      string        `json:"currency"`
	Method        PaymentMethod `json:"method"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id,omitempty"`
	CreatedAt     time.Time     `json:"created_at"`
	UpdatedAt     time.Time     `json:"updated_at"`
	Error         string        `json:"error,omitempty"`
}

type PaymentRequest struct {
	OrderID    string        `json:"order_id"`
	CustomerID string        `json:"customer_id"`
	Amount     float64       `json:"amount"`
	Currency   string        `json:"currency"`
	Method     PaymentMethod `json:"method"`
}

type PaymentResponse struct {
	PaymentID     string        `json:"payment_id"`
	OrderID       string        `json:"order_id"`
	Status        PaymentStatus `json:"status"`
	TransactionID string        `json:"transaction_id,omitempty"`
	Message       string        `json:"message"`
}

type MockPaymentService struct {
	payments map[string]*Payment
}

func NewMockPaymentService() *MockPaymentService {
	return &MockPaymentService{
		payments: make(map[string]*Payment),
	}
}

func (s *MockPaymentService) ProcessPayment(request PaymentRequest) (*PaymentResponse, error) {
	payment := &Payment{
		ID:         fmt.Sprintf("PAY-%s", request.OrderID),
		OrderID:    request.OrderID,
		CustomerID: request.CustomerID,
		Amount:     request.Amount,
		Currency:   request.Currency,
		Method:     request.Method,
		Status:     PaymentStatusProcessing,
		CreatedAt:  time.Now(),
		UpdatedAt:  time.Now(),
	}

	time.Sleep(100 * time.Millisecond)

	if request.Amount > 1000 {
		payment.Status = PaymentStatusFailed
		payment.Error = "Amount exceeds limit"
		return &PaymentResponse{
			PaymentID: payment.ID,
			OrderID:   request.OrderID,
			Status:    PaymentStatusFailed,
			Message:   payment.Error,
		}, fmt.Errorf("payment failed: %s", payment.Error)
	}

	if request.Method == PaymentMethodPayPal && request.Amount < 10 {
		payment.Status = PaymentStatusFailed
		payment.Error = "PayPal minimum amount is $10"
		return &PaymentResponse{
			PaymentID: payment.ID,
			OrderID:   request.OrderID,
			Status:    PaymentStatusFailed,
			Message:   payment.Error,
		}, fmt.Errorf("payment failed: %s", payment.Error)
	}

	payment.Status = PaymentStatusCompleted
	payment.TransactionID = fmt.Sprintf("TXN-%s-%d", request.OrderID, time.Now().Unix())
	payment.UpdatedAt = time.Now()

	s.payments[payment.ID] = payment

	return &PaymentResponse{
		PaymentID:     payment.ID,
		OrderID:       request.OrderID,
		Status:        PaymentStatusCompleted,
		TransactionID: payment.TransactionID,
		Message:       "Payment processed successfully",
	}, nil
}

func (s *MockPaymentService) GetPayment(paymentID string) (*Payment, error) {
	payment, exists := s.payments[paymentID]
	if !exists {
		return nil, fmt.Errorf("payment %s not found", paymentID)
	}
	return payment, nil
}
