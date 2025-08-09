package inventory

import (
	"fmt"
	"time"
)

type ProductAvailability struct {
	ProductID   string    `json:"product_id"`
	Available   bool      `json:"available"`
	Quantity    int       `json:"quantity"`
	LastUpdated time.Time `json:"last_updated"`
}

type InventoryCheckRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type InventoryCheckResponse struct {
	ProductID   string `json:"product_id"`
	Available   bool   `json:"available"`
	Requested   int    `json:"requested"`
	AvailableQty int   `json:"available_qty"`
	Message     string `json:"message"`
}

type MockInventoryService struct {
	products map[string]*ProductAvailability
}

func NewMockInventoryService() *MockInventoryService {
	service := &MockInventoryService{
		products: make(map[string]*ProductAvailability),
	}
	
	service.products["PROD-001"] = &ProductAvailability{
		ProductID:   "PROD-001",
		Available:   true,
		Quantity:    50,
		LastUpdated: time.Now(),
	}
	service.products["PROD-002"] = &ProductAvailability{
		ProductID:   "PROD-002",
		Available:   true,
		Quantity:    25,
		LastUpdated: time.Now(),
	}
	service.products["PROD-003"] = &ProductAvailability{
		ProductID:   "PROD-003",
		Available:   false,
		Quantity:    0,
		LastUpdated: time.Now(),
	}
	
	return service
}

func (s *MockInventoryService) CheckAvailability(request InventoryCheckRequest) (*InventoryCheckResponse, error) {
	product, exists := s.products[request.ProductID]
	if !exists {
		return &InventoryCheckResponse{
			ProductID: request.ProductID,
			Available: false,
			Requested: request.Quantity,
			AvailableQty: 0,
			Message: "Product not found",
		}, fmt.Errorf("product %s not found", request.ProductID)
	}

	if !product.Available {
		return &InventoryCheckResponse{
			ProductID: request.ProductID,
			Available: false,
			Requested: request.Quantity,
			AvailableQty: product.Quantity,
			Message: "Product is not available",
		}, nil
	}

	if product.Quantity < request.Quantity {
		return &InventoryCheckResponse{
			ProductID: request.ProductID,
			Available: false,
			Requested: request.Quantity,
			AvailableQty: product.Quantity,
			Message: fmt.Sprintf("Insufficient quantity. Requested: %d, Available: %d", request.Quantity, product.Quantity),
		}, nil
	}

	return &InventoryCheckResponse{
		ProductID: request.ProductID,
		Available: true,
		Requested: request.Quantity,
		AvailableQty: product.Quantity,
		Message: "Product is available",
	}, nil
}
