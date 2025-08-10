package inventory

import "errors"

var (
	ErrInvalidInput = errors.New("invalid input")
	ErrOutOfStock   = errors.New("out of stock")
	ErrInvalidState = errors.New("invalid state transition")
)