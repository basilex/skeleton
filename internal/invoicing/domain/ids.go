package domain

import (
	"github.com/basilex/skeleton/pkg/uuid"
)

type InvoiceID uuid.UUID

func NewInvoiceID() InvoiceID {
	return InvoiceID(uuid.NewV7())
}

func ParseInvoiceID(s string) (InvoiceID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return InvoiceID{}, err
	}
	return InvoiceID(id), nil
}

func (id InvoiceID) String() string {
	return uuid.UUID(id).String()
}

func (id InvoiceID) IsZero() bool {
	return uuid.UUID(id) == uuid.UUID{}
}

type InvoiceLineID uuid.UUID

func NewInvoiceLineID() InvoiceLineID {
	return InvoiceLineID(uuid.NewV7())
}

func ParseInvoiceLineID(s string) (InvoiceLineID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return InvoiceLineID{}, err
	}
	return InvoiceLineID(id), nil
}

func (id InvoiceLineID) String() string {
	return uuid.UUID(id).String()
}

type PaymentID uuid.UUID

func NewPaymentID() PaymentID {
	return PaymentID(uuid.NewV7())
}

func ParsePaymentID(s string) (PaymentID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return PaymentID{}, err
	}
	return PaymentID(id), nil
}

func (id PaymentID) String() string {
	return uuid.UUID(id).String()
}
