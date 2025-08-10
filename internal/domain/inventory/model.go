package inventory

import "time"

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

type Reservation struct {
	ID         string    `json:"id"`
	OrderID    string    `json:"order_id"`
	ProductID  string    `json:"product_id"`
	Quantity   int       `json:"quantity"`
	ExpiresAt  time.Time `json:"expires_at"`
	CreatedAt  time.Time `json:"created_at"`
}

type CheckRequest struct {
	OrderID string      `json:"order_id"`
	Items   []CheckItem `json:"items"`
}

type CheckItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

type CheckResponse struct {
	Available        bool                `json:"available"`
	UnavailableItems []UnavailableItem   `json:"unavailable_items,omitempty"`
	ReservationID    string              `json:"reservation_id,omitempty"`
}

type UnavailableItem struct {
	ProductID        string `json:"product_id"`
	RequestedQuantity int   `json:"requested_quantity"`
	AvailableQuantity int   `json:"available_quantity"`
}

type ReserveRequest struct {
	OrderID string        `json:"order_id"`
	Items   []ReserveItem `json:"items"`
}

type ReserveItem struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (p *Product) IsAvailable(quantity int) bool {
	return p.Available >= quantity
}

func (p *Product) CanReserve(quantity int) bool {
	return p.Available-p.Reserved >= quantity
}

func (p *Product) Reserve(quantity int) error {
	if !p.CanReserve(quantity) {
		return NewInsufficientStockError(p.ID, quantity, p.Available-p.Reserved)
	}
	p.Reserved += quantity
	p.UpdatedAt = time.Now()
	return nil
}

func (p *Product) ReleaseReservation(quantity int) {
	if p.Reserved >= quantity {
		p.Reserved -= quantity
	} else {
		p.Reserved = 0
	}
	p.UpdatedAt = time.Now()
}

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

func (r *Reservation) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}