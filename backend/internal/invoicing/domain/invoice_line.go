package domain

import (
	"errors"

	"github.com/basilex/skeleton/pkg/money"
)

type InvoiceLine struct {
	id          InvoiceLineID
	invoiceID   InvoiceID
	description string
	quantity    float64
	unitPrice   money.Money
	unit        string
	discount    money.Money
	total       money.Money
}

func NewInvoiceLine(
	invoiceID InvoiceID,
	description string,
	quantity float64,
	unitPrice money.Money,
	unit string,
	discount money.Money,
) (*InvoiceLine, error) {
	if description == "" {
		return nil, errors.New("description cannot be empty")
	}
	if quantity <= 0 {
		return nil, ErrInvalidLineQuantity
	}
	if unitPrice.IsNegative() {
		return nil, ErrInvalidLinePrice
	}
	if discount.IsNegative() {
		return nil, errors.New("discount cannot be negative")
	}

	total, err := unitPrice.Multiply(quantity)
	if err != nil {
		return nil, err
	}
	total, err = total.Subtract(discount)
	if err != nil {
		return nil, err
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

func (l *InvoiceLine) GetUnitPrice() money.Money {
	return l.unitPrice
}

func (l *InvoiceLine) GetUnit() string {
	return l.unit
}

func (l *InvoiceLine) GetDiscount() money.Money {
	return l.discount
}

func (l *InvoiceLine) GetTotal() money.Money {
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

func (l *InvoiceLine) UpdateUnitPrice(price money.Money) error {
	if price.IsNegative() {
		return ErrInvalidLinePrice
	}
	l.unitPrice = price
	l.recalculateTotal()
	return nil
}

func (l *InvoiceLine) UpdateDiscount(discount money.Money) error {
	if discount.IsNegative() {
		return errors.New("discount cannot be negative")
	}
	l.discount = discount
	l.recalculateTotal()
	return nil
}

func (l *InvoiceLine) recalculateTotal() {
	total, err := l.unitPrice.Multiply(l.quantity)
	if err != nil {
		l.total = money.Money{}
		return
	}
	if !l.discount.IsZero() {
		total, err = total.Subtract(l.discount)
		if err != nil {
			l.total = money.Money{}
			return
		}
	}
	l.total = total
}

func RestoreInvoiceLine(
	id InvoiceLineID,
	invoiceID InvoiceID,
	description string,
	quantity float64,
	unitPrice money.Money,
	unit string,
	discount money.Money,
	total money.Money,
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
