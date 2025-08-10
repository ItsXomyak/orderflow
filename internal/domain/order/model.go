package order

type Item struct {
	SKU      SKU
	Quantity Quantity
	Price    Money 
}

type Snapshot struct {
	ID     OrderID
	Status Status
	Items  []Item
	Total  Money
}
