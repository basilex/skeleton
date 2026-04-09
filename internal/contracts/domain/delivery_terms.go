package domain

import (
	"encoding/json"
	"fmt"
)

type DeliveryType string

const (
	DeliveryTypePickup   DeliveryType = "pickup"
	DeliveryTypeDelivery DeliveryType = "delivery"
	DeliveryTypeDigital  DeliveryType = "digital"
)

func (dt DeliveryType) String() string {
	return string(dt)
}

func ParseDeliveryType(s string) (DeliveryType, error) {
	switch DeliveryType(s) {
	case DeliveryTypePickup, DeliveryTypeDelivery, DeliveryTypeDigital:
		return DeliveryType(s), nil
	default:
		return "", fmt.Errorf("invalid delivery type: %s", s)
	}
}

type DeliveryTerms struct {
	DeliveryType     DeliveryType `json:"delivery_type"`
	EstimatedDays    int          `json:"estimated_days"`
	ShippingCost     float64      `json:"shipping_cost"`
	Insurance        bool         `json:"insurance"`
	ShippingCurrency string       `json:"shipping_currency"`
}

func NewDeliveryTerms(deliveryType DeliveryType, estimatedDays int) DeliveryTerms {
	return DeliveryTerms{
		DeliveryType:  deliveryType,
		EstimatedDays: estimatedDays,
	}
}

func (dt DeliveryTerms) Validate() error {
	if dt.EstimatedDays < 0 {
		return fmt.Errorf("estimated days cannot be negative")
	}
	return nil
}

func (dt DeliveryTerms) ToJSON() (json.RawMessage, error) {
	data, err := json.Marshal(dt)
	if err != nil {
		return nil, fmt.Errorf("marshal delivery terms: %w", err)
	}
	return json.RawMessage(data), nil
}

func DeliveryTermsFromJSON(data json.RawMessage) (DeliveryTerms, error) {
	var dt DeliveryTerms
	if len(data) == 0 {
		return DeliveryTerms{}, nil
	}
	if err := json.Unmarshal(data, &dt); err != nil {
		return DeliveryTerms{}, fmt.Errorf("unmarshal delivery terms: %w", err)
	}
	return dt, nil
}
