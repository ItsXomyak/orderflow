package order

type OrderCreated struct {
	OrderID OrderID
}
type InventoryReserved struct {
	OrderID OrderID
}
type PaymentSucceeded struct {
	OrderID     OrderID
	ProviderRef string
}
type OrderCancelled struct {
	OrderID OrderID
	Reason  string
}
type OrderCompleted struct {
	OrderID OrderID
}
