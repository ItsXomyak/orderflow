package payment

type ID string

type Payment struct {
	ID          ID
	OrderID     string
	Amount      int64
	Status      Status
	Provider    string
	ProviderRef string
}
