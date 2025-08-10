package order

type Status string

const (
	StatusCreated           Status = "CREATED"
	StatusInventoryReserved Status = "INVENTORY_RESERVED"
	StatusPaid              Status = "PAID"
	StatusCancelled         Status = "CANCELLED"
	StatusFailed            Status = "FAILED"
	StatusCompleted         Status = "COMPLETED"
)
