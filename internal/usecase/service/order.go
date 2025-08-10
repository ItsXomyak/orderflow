package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"orderflow/internal/domain/order"
)

type OrderStatistics struct {
	TotalOrders     int     `json:"total_orders"`
	CompletedOrders int     `json:"completed_orders"`
	FailedOrders    int     `json:"failed_orders"`
	CancelledOrders int     `json:"cancelled_orders"`
	TotalAmount     float64 `json:"total_amount"`
	AverageAmount   float64 `json:"average_amount"`
}

type OrderService struct {
	orderRepo order.Repository
}

func NewOrderService(orderRepo order.Repository) *OrderService {
	return &OrderService{
		orderRepo: orderRepo,
	}
}

func (s *OrderService) Create(ctx context.Context, req *order.CreateRequest) (*order.Order, error) {
	if req.CustomerID == "" {
		return nil, order.NewValidationError("customer_id is required")
	}

	if len(req.Items) == 0 {
		return nil, order.NewValidationError("order must have at least one item")
	}

	for i, item := range req.Items {
		if item.ProductID == "" {
			return nil, order.NewValidationError("product_id is required for item " + string(rune(i)))
		}
		if item.Quantity <= 0 {
			return nil, order.NewValidationError("quantity must be positive for item " + string(rune(i)))
		}
		if item.Price < 0 {
			return nil, order.NewValidationError("price cannot be negative for item " + string(rune(i)))
		}
	}

	newOrder := order.NewOrder(req.CustomerID, req.Items)
	newOrder.ID = uuid.New().String()

	if err := newOrder.Validate(); err != nil {
		return nil, err
	}

	if err := s.orderRepo.Create(ctx, newOrder); err != nil {
		return nil, err
	}

	return newOrder, nil
}

func (s *OrderService) GetByID(ctx context.Context, id string) (*order.Order, error) {
	if id == "" {
		return nil, order.NewValidationError("order_id is required")
	}

	orderEntity, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if orderEntity == nil {
		return nil, order.NewNotFoundError(id)
	}

	return orderEntity, nil
}

func (s *OrderService) Cancel(ctx context.Context, id string) error {
	if id == "" {
		return order.NewValidationError("order_id is required")
	}

	orderEntity, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if orderEntity == nil {
		return order.NewNotFoundError(id)
	}

	if err := orderEntity.Cancel(); err != nil {
		return err
	}

	return s.orderRepo.Update(ctx, orderEntity)
}

func (s *OrderService) UpdateStatus(ctx context.Context, id string, status order.Status) error {
	if id == "" {
		return order.NewValidationError("order_id is required")
	}

	orderEntity, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if orderEntity == nil {
		return order.NewNotFoundError(id)
	}

	if !s.isValidStatusTransition(orderEntity.Status, status) {
		return order.NewStatusTransitionError(orderEntity.Status, status)
	}

	return s.orderRepo.UpdateStatus(ctx, id, status)
}

func (s *OrderService) SetFailure(ctx context.Context, id string, reason string) error {
	if id == "" {
		return order.NewValidationError("order_id is required")
	}

	if reason == "" {
		reason = "Unknown error"
	}

	return s.orderRepo.SetFailure(ctx, id, reason)
}

func (s *OrderService) Complete(ctx context.Context, id string, paymentID string) error {
	if id == "" {
		return order.NewValidationError("order_id is required")
	}

	orderEntity, err := s.orderRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if orderEntity == nil {
		return order.NewNotFoundError(id)
	}

	orderEntity.PaymentID = paymentID
	orderEntity.UpdateStatus(order.StatusCompleted)

	return s.orderRepo.Update(ctx, orderEntity)
}

func (s *OrderService) GetByCustomerID(ctx context.Context, customerID string) ([]*order.Order, error) {
	if customerID == "" {
		return nil, order.NewValidationError("customer_id is required")
	}

	return s.orderRepo.GetByCustomerID(ctx, customerID)
}

func (s *OrderService) List(ctx context.Context, offset, limit int) ([]*order.Order, error) {
	if offset < 0 {
		offset = 0
	}
	if limit <= 0 || limit > 100 {
		limit = 20 // значение по умолчанию
	}

	return s.orderRepo.List(ctx, offset, limit)
}

func (s *OrderService) isValidStatusTransition(from, to order.Status) bool {
	validTransitions := map[order.Status][]order.Status{
		order.StatusPending: {
			order.StatusValidating,
			order.StatusFailed,
			order.StatusCancelled,
		},
		order.StatusValidating: {
			order.StatusPayment,
			order.StatusFailed,
			order.StatusCancelled,
		},
		order.StatusPayment: {
			order.StatusCompleted,
			order.StatusFailed,
			order.StatusCancelled,
		},
		order.StatusCompleted: {
			// завершенный заказ нельзя изменить
		},
		order.StatusFailed: {
			// неудачный заказ нельзя изменить
		},
		order.StatusCancelled: {
			// отмененный заказ нельзя изменить
		},
	}

	allowedStatuses, exists := validTransitions[from]
	if !exists {
		return false
	}

	for _, allowedStatus := range allowedStatuses {
		if allowedStatus == to {
			return true
		}
	}

	return false
}

func (s *OrderService) GetOrderStatistics(ctx context.Context, customerID string, from, to time.Time) (*OrderStatistics, error) {
	orders, err := s.orderRepo.GetByCustomerID(ctx, customerID)
	if err != nil {
		return nil, err
	}

	stats := &OrderStatistics{
		TotalOrders:     0,
		CompletedOrders: 0,
		FailedOrders:    0,
		CancelledOrders: 0,
		TotalAmount:     0,
		AverageAmount:   0,
	}

	for _, orderEntity := range orders {
		if !from.IsZero() && orderEntity.CreatedAt.Before(from) {
			continue
		}
		if !to.IsZero() && orderEntity.CreatedAt.After(to) {
			continue
		}

		stats.TotalOrders++
		stats.TotalAmount += orderEntity.TotalAmount

		switch orderEntity.Status {
		case order.StatusCompleted:
			stats.CompletedOrders++
		case order.StatusFailed:
			stats.FailedOrders++
		case order.StatusCancelled:
			stats.CancelledOrders++
		}
	}

	if stats.TotalOrders > 0 {
		stats.AverageAmount = stats.TotalAmount / float64(stats.TotalOrders)
	}

	return stats, nil
}
