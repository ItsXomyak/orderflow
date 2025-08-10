// internal/domain/inventory/entity.go
package inventory

import "time"

// Product представляет товар
type Product struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	SKU         string    `json:"sku"`
	Price       float64   `json:"price"`
	Available   int       `json:"available"`
	Reserved    int       `json:"reserved"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// Reservation представляет резервирование товара
type Reservation struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

// CheckRequest запрос на проверку наличия товаров
type CheckRequest struct {
	OrderID string      `json:"order_id"`
	Items   []CheckItem `json:"items"`
}

// CheckItem элемент для проверки наличия
type CheckItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// CheckResponse ответ проверки наличия товаров
type CheckResponse struct {
	Available        bool                `json:"available"`
	UnavailableItems []UnavailableItem   `json:"unavailable_items,omitempty"`
	ReservationID    string              `json:"reservation_id,omitempty"`
}

// UnavailableItem информация о недоступном товаре
type UnavailableItem struct {
	ProductID        string `json:"product_id"`
	RequestedQuantity int   `json:"requested_quantity"`
	AvailableQuantity int   `json:"available_quantity"`
}

// ReserveRequest запрос на резервирование товаров
type ReserveRequest struct {
	OrderID string        `json:"order_id"`
	Items   []ReserveItem `json:"items"`
}

// ReserveItem элемент для резервирования
type ReserveItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

// IsAvailable проверяет, доступно ли нужное количество товара
func (p *Product) IsAvailable(quantity int) bool {
	return p.Available >= quantity
}

// CanReserve проверяет, можно ли зарезервировать товар
func (p *Product) CanReserve(quantity int) bool {
	return p.Available-p.Reserved >= quantity
}

// Reserve резервирует товар
func (p *Product) Reserve(quantity int) error {
	if !p.CanReserve(quantity) {
		return NewInsufficientStockError(p.ID, quantity, p.Available-p.Reserved)
	}
	p.Reserved += quantity
	p.UpdatedAt = time.Now()
	return nil
}

// ReleaseReservation освобождает резервирование
func (p *Product) ReleaseReservation(quantity int) {
	if p.Reserved >= quantity {
		p.Reserved -= quantity
	} else {
		p.Reserved = 0
	}
	p.UpdatedAt = time.Now()
}

// Sell продает товар (уменьшает доступное количество)
func (p *Product) Sell(quantity int) error {
	if p.Available < quantity {
		return NewInsufficientStockError(p.ID, quantity, p.Available)
	}
	p.Available -= quantity
	if p.Reserved >= quantity {
		p.Reserved -= quantity
	}
	p.UpdatedAt = time.Now()
	return nil
}

// IsExpired проверяет, истекло ли резервирование
func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}