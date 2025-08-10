// internal/domain/payment/service.go
package payment

import (
	"context"
	"time"
)

// Service интерфейс для работы с платежами
type Service interface {
	// ProcessPayment обрабатывает платеж
	ProcessPayment(ctx context.Context, req *Request) (*Response, error)
	
	// GetPayment получает информацию о платеже
	GetPayment(ctx context.Context, paymentID string) (*Payment, error)
	
	// GetPaymentByOrderID получает платеж по ID заказа
	GetPaymentByOrderID(ctx context.Context, orderID string) (*Payment, error)
	
	// RefundPayment возвращает деньги
	RefundPayment(ctx context.Context, req *RefundRequest) error
	
	// CancelPayment отменяет платеж (если он еще в обработке)
	CancelPayment(ctx context.Context, paymentID string) error
}

// Gateway интерфейс для интеграции с внешними платежными системами
type Gateway interface {
	// Charge списывает деньги
	Charge(ctx context.Context, req *Request) (*Response, error)
	
	// Refund возвращает деньги
	Refund(ctx context.Context, transactionID string, amount float64) error
	
	// GetTransaction получает информацию о транзакции
	GetTransaction(ctx context.Context, transactionID string) (*Transaction, error)
}

// Transaction представляет транзакцию во внешней платежной системе
type Transaction struct {
	ID        string    `json:"id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}