package domain

import (
	"fmt"
	"time"
)

type OrderLine struct {
	id        OrderLineID
	orderID   OrderID
	itemID    string
	itemName  string
	quantity  float64
	unit      string
	unitPrice float64
	discount  float64
	total     float64
	metadata  map[string]interface{}
	createdAt time.Time
}

func NewOrderLine(
	orderID OrderID,
	itemID, itemName string,
	quantity float64,
	unit string,
	unitPrice, discount float64,
) (*OrderLine, error) {
	if quantity <= 0 {
		return nil, ErrInvalidQuantity
	}
	if unitPrice < 0 {
		return nil, fmt.Errorf("unit price cannot be negative")
	}
	if discount < 0 {
		return nil, fmt.Errorf("discount cannot be negative")
	}

	total := (quantity * unitPrice) - discount
	if total < 0 {
		total = 0
	}

	return &OrderLine{
		id:        NewOrderLineID(),
		orderID:   orderID,
		itemID:    itemID,
		itemName:  itemName,
		quantity:  quantity,
		unit:      unit,
		unitPrice: unitPrice,
		discount:  discount,
		total:     total,
		metadata:  make(map[string]interface{}),
		createdAt: time.Now().UTC(),
	}, nil
}

func (ol *OrderLine) GetID() OrderLineID      { return ol.id }
func (ol *OrderLine) GetOrderID() OrderID     { return ol.orderID }
func (ol *OrderLine) GetItemID() string       { return ol.itemID }
func (ol *OrderLine) GetItemName() string     { return ol.itemName }
func (ol *OrderLine) GetQuantity() float64    { return ol.quantity }
func (ol *OrderLine) GetUnit() string         { return ol.unit }
func (ol *OrderLine) GetUnitPrice() float64   { return ol.unitPrice }
func (ol *OrderLine) GetDiscount() float64    { return ol.discount }
func (ol *OrderLine) GetTotal() float64       { return ol.total }
func (ol *OrderLine) GetCreatedAt() time.Time { return ol.createdAt }

func ReconstituteOrderLine(
	id OrderLineID,
	orderID OrderID,
	itemID, itemName string,
	quantity float64,
	unit string,
	unitPrice, discount, total float64,
	createdAt time.Time,
) (*OrderLine, error) {
	return &OrderLine{
		id:        id,
		orderID:   orderID,
		itemID:    itemID,
		itemName:  itemName,
		quantity:  quantity,
		unit:      unit,
		unitPrice: unitPrice,
		discount:  discount,
		total:     total,
		metadata:  make(map[string]interface{}),
		createdAt: createdAt,
	}, nil
}

func (ol *OrderLine) UpdateQuantity(quantity float64) error {
	if quantity <= 0 {
		return ErrInvalidQuantity
	}
	ol.quantity = quantity
	ol.total = (quantity * ol.unitPrice) - ol.discount
	return nil
}

type Order struct {
	id          OrderID
	orderNumber string
	status      OrderStatus

	// Parties
	customerID string
	supplierID string
	contractID string

	// Amounts
	subtotal  float64
	taxAmount float64
	discount  float64
	total     float64
	currency  string

	//Lines
	lines []*OrderLine

	// Dates
	orderDate   time.Time
	dueDate     *time.Time
	completedAt *time.Time
	cancelledAt *time.Time

	// Notes
	notes string

	// Metadata
	metadata map[string]interface{}

	// Audit
	createdBy string
	createdAt time.Time
	updatedAt time.Time

	events []DomainEvent
}

func NewOrder(
	orderNumber string,
	customerID, supplierID, contractID string,
	currency, createdBy string,
) (*Order, error) {
	if orderNumber == "" {
		return nil, fmt.Errorf("order number is required")
	}
	if customerID == "" {
		return nil, fmt.Errorf("customer ID is required")
	}
	if supplierID == "" {
		return nil, fmt.Errorf("supplier ID is required")
	}

	now := time.Now().UTC()
	o := &Order{
		id:          NewOrderID(),
		orderNumber: orderNumber,
		status:      OrderStatusDraft,
		customerID:  customerID,
		supplierID:  supplierID,
		contractID:  contractID,
		subtotal:    0,
		taxAmount:   0,
		discount:    0,
		total:       0,
		currency:    currency,
		lines:       make([]*OrderLine, 0),
		orderDate:   now,
		metadata:    make(map[string]interface{}),
		createdBy:   createdBy,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	o.events = append(o.events, OrderCreated{
		OrderID:    o.id,
		CustomerID: o.customerID,
		SupplierID: o.supplierID,
		Total:      o.total,
		Currency:   o.currency,
		occurredAt: now,
	})

	return o, nil
}

func (o *Order) GetID() OrderID             { return o.id }
func (o *Order) GetOrderNumber() string     { return o.orderNumber }
func (o *Order) GetStatus() OrderStatus     { return o.status }
func (o *Order) GetCustomerID() string      { return o.customerID }
func (o *Order) GetSupplierID() string      { return o.supplierID }
func (o *Order) GetContractID() string      { return o.contractID }
func (o *Order) GetSubtotal() float64       { return o.subtotal }
func (o *Order) GetTaxAmount() float64      { return o.taxAmount }
func (o *Order) GetDiscount() float64       { return o.discount }
func (o *Order) GetTotal() float64          { return o.total }
func (o *Order) GetCurrency() string        { return o.currency }
func (o *Order) GetLines() []*OrderLine     { return o.lines }
func (o *Order) GetOrderDate() time.Time    { return o.orderDate }
func (o *Order) GetDueDate() *time.Time     { return o.dueDate }
func (o *Order) GetCompletedAt() *time.Time { return o.completedAt }
func (o *Order) GetCancelledAt() *time.Time { return o.cancelledAt }
func (o *Order) GetNotes() string           { return o.notes }
func (o *Order) GetCreatedBy() string       { return o.createdBy }
func (o *Order) GetCreatedAt() time.Time    { return o.createdAt }
func (o *Order) GetUpdatedAt() time.Time    { return o.updatedAt }

func (o *Order) AddLine(line *OrderLine) error {
	if o.status != OrderStatusDraft {
		return fmt.Errorf("cannot add lines to order in %s status", o.status)
	}
	o.lines = append(o.lines, line)
	o.recalculateTotals()
	o.updatedAt = time.Now().UTC()
	return nil
}

func (o *Order) RemoveLine(lineID OrderLineID) error {
	if o.status != OrderStatusDraft {
		return fmt.Errorf("cannot remove lines from order in %s status", o.status)
	}

	for i, line := range o.lines {
		if line.GetID() == lineID {
			o.lines = append(o.lines[:i], o.lines[i+1:]...)
			o.recalculateTotals()
			o.updatedAt = time.Now().UTC()
			return nil
		}
	}

	return ErrOrderLineNotFound
}

func (o *Order) recalculateTotals() {
	o.subtotal = 0
	for _, line := range o.lines {
		o.subtotal += line.GetTotal()
	}
	o.total = o.subtotal + o.taxAmount - o.discount
}

func (o *Order) Confirm() error {
	if o.status != OrderStatusDraft && o.status != OrderStatusPending {
		return fmt.Errorf("cannot confirm order in %s status", o.status)
	}
	if len(o.lines) == 0 {
		return fmt.Errorf("cannot confirm order without lines")
	}

	oldStatus := o.status
	o.status = OrderStatusConfirmed
	o.updatedAt = time.Now().UTC()

	o.events = append(o.events, OrderStatusChanged{
		OrderID:    o.id,
		OldStatus:  oldStatus,
		NewStatus:  o.status,
		occurredAt: o.updatedAt,
	})

	o.events = append(o.events, OrderConfirmed{
		OrderID:     o.id,
		CustomerID:  o.customerID,
		SupplierID:  o.supplierID,
		WarehouseID: "", // TODO: Get from order or configuration
		Lines:       o.convertLinesToConfirmedLines(),
		Total:       o.total,
		Currency:    o.currency,
		occurredAt:  o.updatedAt,
	})

	return nil
}

func (o *Order) Complete() error {
	if o.status != OrderStatusConfirmed && o.status != OrderStatusProcessing {
		return ErrOrderCannotComplete
	}

	oldStatus := o.status
	now := time.Now().UTC()
	o.status = OrderStatusCompleted
	o.completedAt = &now
	o.updatedAt = now

	o.events = append(o.events, OrderStatusChanged{
		OrderID:    o.id,
		OldStatus:  oldStatus,
		NewStatus:  o.status,
		occurredAt: now,
	})

	o.events = append(o.events, OrderCompleted{
		OrderID:    o.id,
		CustomerID: o.customerID,
		Total:      o.total,
		occurredAt: now,
	})

	return nil
}

func (o *Order) Cancel(reason string) error {
	if o.status == OrderStatusCompleted || o.status == OrderStatusCancelled {
		return ErrOrderCannotCancel
	}

	oldStatus := o.status
	now := time.Now().UTC()
	o.status = OrderStatusCancelled
	o.cancelledAt = &now
	o.notes = reason
	o.updatedAt = now

	o.events = append(o.events, OrderStatusChanged{
		OrderID:    o.id,
		OldStatus:  oldStatus,
		NewStatus:  o.status,
		occurredAt: now,
	})

	o.events = append(o.events, OrderCancelled{
		OrderID:    o.id,
		CustomerID: o.customerID,
		Reason:     reason,
		occurredAt: now,
	})

	return nil
}

func (o *Order) convertLinesToConfirmedLines() []OrderConfirmedLine {
	lines := make([]OrderConfirmedLine, len(o.lines))
	for i, line := range o.lines {
		lines[i] = OrderConfirmedLine{
			ItemID:    line.itemID,
			ItemName:  line.itemName,
			Quantity:  line.quantity,
			Unit:      line.unit,
			UnitPrice: line.unitPrice,
			Discount:  line.discount,
			Total:     line.total,
		}
	}
	return lines
}

func (o *Order) PullEvents() []DomainEvent {
	events := o.events
	o.events = make([]DomainEvent, 0)
	return events
}

func ReconstituteOrder(
	id OrderID,
	orderNumber string,
	status OrderStatus,
	customerID, supplierID, contractID string,
	subtotal, taxAmount, discount, total float64,
	currency string,
	lines []*OrderLine,
	orderDate time.Time,
	dueDate, completedAt, cancelledAt *time.Time,
	notes string,
	metadata map[string]interface{},
	createdBy string,
	createdAt, updatedAt time.Time,
) (*Order, error) {
	return &Order{
		id:          id,
		orderNumber: orderNumber,
		status:      status,
		customerID:  customerID,
		supplierID:  supplierID,
		contractID:  contractID,
		subtotal:    subtotal,
		taxAmount:   taxAmount,
		discount:    discount,
		total:       total,
		currency:    currency,
		lines:       lines,
		orderDate:   orderDate,
		dueDate:     dueDate,
		completedAt: completedAt,
		cancelledAt: cancelledAt,
		notes:       notes,
		metadata:    metadata,
		createdBy:   createdBy,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}, nil
}
