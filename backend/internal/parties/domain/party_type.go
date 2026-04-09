package domain

import "fmt"

type PartyType string

const (
	PartyTypeCustomer PartyType = "customer"
	PartyTypeSupplier PartyType = "supplier"
	PartyTypePartner  PartyType = "partner"
	PartyTypeEmployee PartyType = "employee"
)

func (pt PartyType) String() string {
	return string(pt)
}

func ParsePartyType(s string) (PartyType, error) {
	switch PartyType(s) {
	case PartyTypeCustomer, PartyTypeSupplier, PartyTypePartner, PartyTypeEmployee:
		return PartyType(s), nil
	default:
		return "", fmt.Errorf("invalid party type: %s", s)
	}
}
