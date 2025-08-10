package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"orderflow/internal/domain/inventory"
	"orderflow/pkg/logger"
)

type InventoryService struct {
	inventoryRepo inventory.Repository
}

func NewInventoryService(inventoryRepo inventory.Repository) *InventoryService {
	return &InventoryService{
		inventoryRepo: inventoryRepo,
	}
}

func (service *InventoryService) CheckAvailability(ctx context.Context, req *inventory.CheckRequest) (*inventory.CheckResponse, error) {
	logger.Info("Checking inventory availability", "order_id", req.OrderID, "items_count", len(req.Items))

	if req.OrderID == "" {
		return nil, inventory.NewValidationError("order_id is required")
	}

	if len(req.Items) == 0 {
		return nil, inventory.NewValidationError("items are required")
	}

	response := &inventory.CheckResponse{
		Available:        true,
		UnavailableItems: make([]inventory.UnavailableItem, 0),
	}

	for _, item := range req.Items {
		product, err := service.inventoryRepo.GetProduct(ctx, item.ProductID)
		if err != nil {
			if _, ok := err.(*inventory.ProductNotFoundError); ok {
				response.Available = false
				response.UnavailableItems = append(response.UnavailableItems, inventory.UnavailableItem{
					ProductID:         item.ProductID,
					RequestedQuantity: item.Quantity,
					AvailableQuantity: 0,
				})
				continue
			}
			return nil, err
		}

		if !product.IsAvailable(item.Quantity) {
			response.Available = false
			response.UnavailableItems = append(response.UnavailableItems, inventory.UnavailableItem{
				ProductID:         item.ProductID,
				RequestedQuantity: item.Quantity,
				AvailableQuantity: product.Available,
			})
		}
	}

	if response.Available {
		logger.Info("Inventory check passed", "order_id", req.OrderID)
	} else {
		logger.Warn("Inventory check failed", "order_id", req.OrderID, "unavailable_items", response.UnavailableItems)
	}

	return response, nil
}

func (service *InventoryService) ReserveItems(ctx context.Context, req *inventory.ReserveRequest) error {
	logger.Info("Reserving items", "order_id", req.OrderID, "items_count", len(req.Items))

	if req.OrderID == "" {
		return inventory.NewValidationError("order_id is required")
	}

	if len(req.Items) == 0 {
		return inventory.NewValidationError("items are required")
	}

	for _, item := range req.Items {
		product, err := service.inventoryRepo.GetProduct(ctx, item.ProductID)
		if err != nil {
			return err
		}

		if !product.CanReserve(item.Quantity) {
			return inventory.NewInsufficientStockError(item.ProductID, item.Quantity, product.Available-product.Reserved)
		}
	}

	for _, item := range req.Items {
		product, err := service.inventoryRepo.GetProduct(ctx, item.ProductID)
		if err != nil {
			return err
		}

		if err := product.Reserve(item.Quantity); err != nil {
			return err
		}

		if err := service.inventoryRepo.UpdateProduct(ctx, product); err != nil {
			return err
		}

		reservation := &inventory.Reservation{
			ID:        uuid.New().String(),
			OrderID:   req.OrderID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
			ExpiresAt: time.Now().Add(30 * time.Minute), // Резервирование на 30 минут
			CreatedAt: time.Now(),
		}

		if err := service.inventoryRepo.CreateReservation(ctx, reservation); err != nil {
			product.ReleaseReservation(item.Quantity)
			service.inventoryRepo.UpdateProduct(ctx, product)
			return err
		}
	}

	logger.Info("Items reserved successfully", "order_id", req.OrderID)
	return nil
}

func (service *InventoryService) ReleaseReservation(ctx context.Context, orderID string) error {
	logger.Info("Releasing reservation", "order_id", orderID)

	if orderID == "" {
		return inventory.NewValidationError("order_id is required")
	}

	reservation, err := service.inventoryRepo.GetReservationByOrderID(ctx, orderID)
	if err != nil {
		if _, ok := err.(*inventory.ReservationNotFoundError); ok {
			logger.Warn("Reservation not found", "order_id", orderID)
			return nil // Не считаем это ошибкой
		}
		return err
	}

	product, err := service.inventoryRepo.GetProduct(ctx, reservation.ProductID)
	if err != nil {
		return err
	}

	product.ReleaseReservation(reservation.Quantity)
	if err := service.inventoryRepo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	if err := service.inventoryRepo.DeleteReservation(ctx, orderID); err != nil {
		return err
	}

	logger.Info("Reservation released", "order_id", orderID, "reservation_id", reservation.ID)
	return nil
}

func (service *InventoryService) ConfirmReservation(ctx context.Context, orderID string) error {
	logger.Info("Confirming reservation", "order_id", orderID)

	if orderID == "" {
		return inventory.NewValidationError("order_id is required")
	}

	reservation, err := service.inventoryRepo.GetReservationByOrderID(ctx, orderID)
	if err != nil {
		return err
	}

	product, err := service.inventoryRepo.GetProduct(ctx, reservation.ProductID)
	if err != nil {
		return err
	}

	if err := product.Sell(reservation.Quantity); err != nil {
		return err
	}

	if err := service.inventoryRepo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	if err := service.inventoryRepo.DeleteReservation(ctx, orderID); err != nil {
		return err
	}

	logger.Info("Reservation confirmed", "order_id", orderID, "reservation_id", reservation.ID)
	return nil
}

func (service *InventoryService) GetProduct(ctx context.Context, productID string) (*inventory.Product, error) {
	if productID == "" {
		return nil, inventory.NewValidationError("product_id is required")
	}

	return service.inventoryRepo.GetProduct(ctx, productID)
}

func (service *InventoryService) UpdateStock(ctx context.Context, productID string, quantity int) error {
	logger.Info("Updating stock", "product_id", productID, "quantity", quantity)

	if productID == "" {
		return inventory.NewValidationError("product_id is required")
	}

	product, err := service.inventoryRepo.GetProduct(ctx, productID)
	if err != nil {
		return err
	}

	product.Available = quantity
	product.UpdatedAt = time.Now()

	if err := service.inventoryRepo.UpdateProduct(ctx, product); err != nil {
		return err
	}

	logger.Info("Stock updated", "product_id", productID, "new_quantity", quantity)
	return nil
}

func (service *InventoryService) GetProducts(ctx context.Context) ([]*inventory.Product, error) {
	return service.inventoryRepo.GetProducts(ctx)
}

func (service *InventoryService) CreateProduct(ctx context.Context, product *inventory.Product) error {
	if product.ID == "" {
		product.ID = uuid.New().String()
	}
	if product.CreatedAt.IsZero() {
		product.CreatedAt = time.Now()
	}
	if product.UpdatedAt.IsZero() {
		product.UpdatedAt = time.Now()
	}

	return service.inventoryRepo.CreateProduct(ctx, product)
}

func (service *InventoryService) CleanupExpiredReservations(ctx context.Context) error {
	logger.Info("Cleaning up expired reservations")

	expiredReservations, err := service.inventoryRepo.GetExpiredReservations(ctx)
	if err != nil {
		return err
	}

	for _, reservation := range expiredReservations {
		logger.Info("Cleaning up expired reservation", 
			"reservation_id", reservation.ID, 
			"order_id", reservation.OrderID)

		product, err := service.inventoryRepo.GetProduct(ctx, reservation.ProductID)
		if err != nil {
			logger.Error("Failed to get product for expired reservation", 
				"error", err, "product_id", reservation.ProductID)
			continue
		}

		product.ReleaseReservation(reservation.Quantity)
		if err := service.inventoryRepo.UpdateProduct(ctx, product); err != nil {
			logger.Error("Failed to update product for expired reservation", 
				"error", err, "product_id", reservation.ProductID)
			continue
		}

		if err := service.inventoryRepo.DeleteReservation(ctx, reservation.OrderID); err != nil {
			logger.Error("Failed to delete expired reservation", 
				"error", err, "reservation_id", reservation.ID)
		}
	}

	logger.Info("Expired reservations cleanup completed", "count", len(expiredReservations))
	return nil
}