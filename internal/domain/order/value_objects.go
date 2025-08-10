package order

import "errors"

type OrderID string
type SKU string

type Quantity int
func NewQuantity(v int) (Quantity, error) {
	if v <= 0 { return 0, errors.New("quantity must be > 0") }
	return Quantity(v), nil
}

type Money int64 
func NewMoney(v int64) (Money, error) {
	if v < 0 { return 0, errors.New("money must be >= 0") }
	return Money(v), nil
}
