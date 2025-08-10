package payment

import "errors"

var (
	ErrDeclined     = errors.New("payment declined")
	ErrInvalidInput = errors.New("invalid payment input")
)
