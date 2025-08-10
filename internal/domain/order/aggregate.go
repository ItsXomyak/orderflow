package order

type Order struct {
	id     OrderID
	status Status
	items  []Item
	total  Money
	events []any // доменные события, типы в events.go
}

func New(id OrderID, items []Item) *Order {
	var total Money
	for _, it := range items {
		total += it.Price * Money(it.Quantity)
	}
	return &Order{id: id, status: StatusCreated, items: items, total: total}
}

func (o *Order) ID() OrderID       { return o.id }
func (o *Order) Status() Status    { return o.status }
func (o *Order) Total() Money      { return o.total }
func (o *Order) Items() []Item     { return append([]Item(nil), o.items...) }
func (o *Order) PullEvents() []any { ev := o.events; o.events = nil; return ev }

func (o *Order) MarkInventoryReserved() error {
	if o.status != StatusCreated {
		return ErrInvalidState
	}
	o.status = StatusInventoryReserved
	o.events = append(o.events, InventoryReserved{OrderID: o.id})
	return nil
}

func (o *Order) MarkPaid(providerRef string) error {
	switch o.status {
	case StatusInventoryReserved:
		o.status = StatusPaid
		o.events = append(o.events, PaymentSucceeded{OrderID: o.id, ProviderRef: providerRef})
		return nil
	case StatusPaid, StatusCompleted:
		return ErrAlreadyPaid
	default:
		return ErrInvalidState
	}
}

func (o *Order) Cancel(reason string) error {
	if o.status == StatusPaid || o.status == StatusCompleted {
		return ErrAlreadyClosed
	}
	o.status = StatusCancelled
	o.events = append(o.events, OrderCancelled{OrderID: o.id, Reason: reason})
	return nil
}

func (o *Order) Complete() error {
	if o.status != StatusPaid {
		return ErrInvalidState
	}
	o.status = StatusCompleted
	o.events = append(o.events, OrderCompleted{OrderID: o.id})
	return nil
}
