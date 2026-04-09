package domain

import "fmt"

type PartyStatus string

const (
	PartyStatusActive      PartyStatus = "active"
	PartyStatusInactive    PartyStatus = "inactive"
	PartyStatusBlacklisted PartyStatus = "blacklisted"
)

func (ps PartyStatus) String() string {
	return string(ps)
}

func ParsePartyStatus(s string) (PartyStatus, error) {
	switch PartyStatus(s) {
	case PartyStatusActive, PartyStatusInactive, PartyStatusBlacklisted:
		return PartyStatus(s), nil
	default:
		return "", fmt.Errorf("invalid party status: %s", s)
	}
}
