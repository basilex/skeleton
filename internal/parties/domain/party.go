package domain

import "time"

type Party interface {
	GetID() PartyID
	GetType() PartyType
	GetName() string
	GetContactInfo() ContactInfo
	GetTaxID() string
	GetStatus() PartyStatus
	GetCreatedAt() time.Time
	GetUpdatedAt() time.Time
}

type LoyaltyLevel string

const (
	LoyaltyLevelBronze   LoyaltyLevel = "bronze"
	LoyaltyLevelSilver   LoyaltyLevel = "silver"
	LoyaltyLevelGold     LoyaltyLevel = "gold"
	LoyaltyLevelPlatinum LoyaltyLevel = "platinum"
)

func (ll LoyaltyLevel) String() string {
	return string(ll)
}

type SupplierRating struct {
	Quality      float64 `json:"quality"`
	DeliveryTime float64 `json:"delivery_time"`
	Price        float64 `json:"price"`
	Overall      float64 `json:"overall"`
}
