package order

import "errors"

var (
	ErrInvalidState   = errors.New("invalid state transition")
	ErrOutOfStock     = errors.New("out of stock")
	ErrAlreadyPaid    = errors.New("order already paid")
	ErrAlreadyClosed  = errors.New("order already closed")
)
