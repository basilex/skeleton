package domain

import (
	"errors"
)

type InvoiceLine struct {
	id          InvoiceLineID
	invoiceID   InvoiceID
	description string
	quantity    float64
	unitPrice   float64
	unit        string
	discount    float64
	total       float64
}

func NewInvoiceLine(
	invoiceID InvoiceID,
	description string,
	quantity float64,
	unitPrice float64,
	unit string,
	discount float64,
) (*InvoiceLine, error) {
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if quantity <= 0 {
		return nil, ErrInvalidLineQuantity
	}
	if unitPrice < 0 {
		return nil, ErrInvalidLinePrice
	}
	if discount < 0 {
		return nil, errors.New("discount cannot be negative")
	}

	total := quantity * unitPrice
	if discount > 0 {
		total = total - discount
		if total < 0 {
			total = 0
		}
	}

	return &InvoiceLine{
		id:          NewInvoiceLineID(),
		invoiceID:   invoiceID,
		description: description,
		quantity:    quantity,
		unitPrice:   unitPrice,
		unit:        unit,
		discount:    discount,
		total:       total,
	}, nil
}

func (l *InvoiceLine) GetID() InvoiceLineID {
	return l.id
}

func (l *InvoiceLine) GetInvoiceID() InvoiceID {
	return l.invoiceID
}

func (l *InvoiceLine) GetDescription() string {
	return l.description
}

func (l *InvoiceLine) GetQuantity() float64 {
	return l.quantity
}

func (l *InvoiceLine) GetUnitPrice() float64 {
	return l.unitPrice
}

func (l *InvoiceLine) GetUnit() string {
	return l.unit
}

func (l *InvoiceLine) GetDiscount() float64 {
	return l.discount
}

func (l *InvoiceLine) GetTotal() float64 {
	return l.total
}

func (l *InvoiceLine) UpdateDescription(desc string) {
	l.description = desc
}

func (l *InvoiceLine) UpdateQuantity(qty float64) error {
	if qty <= 0 {
		return ErrInvalidLineQuantity
	}
	l.quantity = qty
	l.recalculateTotal()
	return nil
}

func (l *InvoiceLine) UpdateUnitPrice(price float64) error {
	if price < 0 {
		return ErrInvalidLinePrice
	}
	l.unitPrice = price
	l.recalculateTotal()
	return nil
}

func (l *InvoiceLine) UpdateDiscount(discount float64) error {
	if discount < 0 {
		return errors.New("discount cannot be negative")
	}
	l.discount = discount
	l.recalculateTotal()
	return nil
}

func (l *InvoiceLine) recalculateTotal() {
	l.total = l.quantity * l.unitPrice
	if l.discount > 0 {
		l.total = l.total - l.discount
		if l.total < 0 {
			l.total = 0
		}
	}
}

func RestoreInvoiceLine(
	id InvoiceLineID,
	invoiceID InvoiceID,
	description string,
	quantity float64,
	unitPrice float64,
	unit string,
	discount float64,
	total float64,
) *InvoiceLine {
	return &InvoiceLine{
		id:          id,
		invoiceID:   invoiceID,
		description: description,
		quantity:    quantity,
		unitPrice:   unitPrice,
		unit:        unit,
		discount:    discount,
		total:       total,
	}
}
