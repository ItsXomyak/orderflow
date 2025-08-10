package payment

type Status string

const (
	StatusInit    Status = "INIT"
	StatusPending Status = "PENDING"
	StatusSuccess Status = "SUCCESS"
	StatusFailed  Status = "FAILED"
)