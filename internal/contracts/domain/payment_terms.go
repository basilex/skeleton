package domain

import (
	"encoding/json"
	"fmt"
)

type PaymentTerms struct {
	PaymentType  PaymentType `json:"payment_type"`
	CreditDays   int         `json:"credit_days"`
	PenaltyRate  float64     `json:"penalty_rate"`
	DiscountRate float64     `json:"discount_rate"`
	Currency     string      `json:"currency"`
}

func NewPaymentTerms(paymentType PaymentType, creditDays int, currency string) (PaymentTerms, error) {
	if paymentType == PaymentTypeCredit && creditDays <= 0 {
		return PaymentTerms{}, fmt.Errorf("credit days must be positive for credit payment type")
	}
	if len(currency) != 3 {
		return PaymentTerms{}, fmt.Errorf("currency must be 3 characters (ISO 4217)")
	}

	return PaymentTerms{
		PaymentType: paymentType,
		CreditDays:  creditDays,
		Currency:    currency,
	}, nil
}

func (pt PaymentTerms) Validate() error {
	if len(pt.Currency) != 3 {
		return ErrInvalidPaymentTerms
	}
	if pt.PaymentType == PaymentTypeCredit && pt.CreditDays <= 0 {
		return ErrInvalidPaymentTerms
	}
	return nil
}

func (pt PaymentTerms) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(pt)
	if err != nil {
		return nil, fmt.Errorf("marshal payment terms: %w", err)
	}
	return json.RawMessage(data), nil
}

func PaymentTermsFromJSON(data json.RawMessage) (PaymentTerms, error) {
	var pt PaymentTerms
	if len(data) == 0 {
		return PaymentTerms{}, nil
	}
	if err := json.Unmarshal(data, &pt); err != nil {
		return PaymentTerms{}, fmt.Errorf("unmarshal payment terms: %w", err)
	}
	return pt, nil
}
