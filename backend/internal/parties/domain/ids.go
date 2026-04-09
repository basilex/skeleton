package domain

import (
	"fmt"

	"github.com/basilex/skeleton/pkg/uuid"
)

type PartyID uuid.UUID

func NewPartyID() PartyID {
	return PartyID(uuid.NewV7())
}

func ParsePartyID(s string) (PartyID, error) {
	u, err := uuid.Parse(s)
	if err != nil {
		return PartyID{}, fmt.Errorf("invalid party id: %w", err)
	}
	return PartyID(u), nil
}

func (id PartyID) String() string {
	return uuid.UUID(id).String()
}
