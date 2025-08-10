package service

import (
	"context"
	"errors"
	"sync"
	"time"

	payment "orderflow/internal/domain/payment"

	"github.com/google/uuid"
)

// Упрощённая, но корректная реализация payment.Service поверх доменных типов
// Хранилище — in-memory (map). Заменяется на БД без изменения интерфейсов.

type paymentService struct {
	gateway  payment.Gateway
	payments map[string]*payment.Payment // paymentID -> Payment
	orders   map[string]string           // orderID -> paymentID
	mu       sync.RWMutex
}

// NewService возвращает реализацию доменного интерфейса payment.Service
func NewService(gw payment.Gateway) payment.Service {
	return &paymentService{
		gateway:  gw,
		payments: make(map[string]*payment.Payment),
		orders:   make(map[string]string),
	}
}

// ProcessPayment обрабатывает платёж через внешний шлюз и фиксирует результат
func (s *paymentService) ProcessPayment(ctx context.Context, req *payment.Request) (*payment.Response, error) {
	if req == nil {
		return nil, errors.New("nil request")
	}
	if req.OrderID == "" {
		return nil, errors.New("order_id is required")
	}
	if req.CustomerID == "" {
		return nil, errors.New("customer_id is required")
	}
	if req.Amount <= 0 {
		return nil, errors.New("amount must be positive")
	}
	if req.Currency == "" {
		return nil, errors.New("currency is required")
	}

	// Создаём доменную сущность (pending)
	p := payment.NewPayment(req)
	p.ID = uuid.NewString()

	// Сохраняем pending перед вызовом шлюза (опционально)
	s.mu.Lock()
	s.payments[p.ID] = p
	s.orders[p.OrderID] = p.ID
	s.mu.Unlock()

	// Делаем списание во внешней системе
	resp, err := s.gateway.Charge(ctx, req)
	if err != nil {
		// Фиксируем fail
		s.mu.Lock()
		p.Fail(err.Error())
		s.mu.Unlock()
		return nil, err
	}

	// Фиксируем результат по Response
	s.mu.Lock()
	defer s.mu.Unlock()
	if resp.Success {
		p.Complete(resp.TransactionID)
		resp.PaymentID = p.ID
	} else {
		p.Fail(resp.ErrorMessage)
		resp.PaymentID = p.ID
	}
	return resp, nil
}

// GetPayment возвращает платёж по его ID
func (s *paymentService) GetPayment(ctx context.Context, paymentID string) (*payment.Payment, error) {
	if paymentID == "" {
		return nil, errors.New("empty paymentID")
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.payments[paymentID]
	if !ok {
		return nil, errors.New("payment not found")
	}
	return p, nil
}

// GetPaymentByOrderID возвращает платёж по ID заказа
func (s *paymentService) GetPaymentByOrderID(ctx context.Context, orderID string) (*payment.Payment, error) {
	if orderID == "" {
		return nil, errors.New("empty orderID")
	}
	s.mu.RLock()
	pid, ok := s.orders[orderID]
	s.mu.RUnlock()
	if !ok {
		return nil, errors.New("payment not found for order")
	}
	return s.GetPayment(ctx, pid)
}

// RefundPayment инициирует возврат; полная сумма если req.Amount == 0
func (s *paymentService) RefundPayment(ctx context.Context, req *payment.RefundRequest) error {
	if req == nil {
		return errors.New("nil refund request")
	}
	p, err := s.GetPayment(ctx, req.PaymentID)
	if err != nil {
		return err
	}
	if !p.CanBeRefunded() {
		return errors.New("payment cannot be refunded in current status")
	}
	amount := req.Amount
	if amount <= 0 {
		amount = p.Amount
	}
	if amount > p.Amount {
		return errors.New("refund amount exceeds payment amount")
	}
	if err := s.gateway.Refund(ctx, p.TransactionID, amount); err != nil {
		return err
	}

	s.mu.Lock()
	p.Status = payment.StatusRefunded
	p.UpdatedAt = time.Now()
	s.mu.Unlock()
	return nil
}

// CancelPayment отменяет платёж, если он ещё не завершён
func (s *paymentService) CancelPayment(ctx context.Context, paymentID string) error {
	p, err := s.GetPayment(ctx, paymentID)
	if err != nil {
		return err
	}
	if p.IsCompleted() || p.IsFailed() || p.IsRefunded() {
		return errors.New("payment cannot be canceled in current status")
	}

	s.mu.Lock()
	p.Status = payment.StatusFailed // либо в домене добавить StatusCanceled и вызывать p.Cancel()
	p.FailureReason = "canceled by user"
	now := time.Now()
	p.ProcessedAt = &now
	p.UpdatedAt = now
	s.mu.Unlock()
	return nil
}
