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
	QualityScore       float64 `json:"quality_score"`
	DeliveryScore      float64 `json:"delivery_score"`
	CommunicationScore float64 `json:"communication_score"`
	OverallScore       float64 `json:"overall_score"`
	RatingCount        int     `json:"rating_count"`
}

type PerformanceLevel string

const (
	PerformanceLevelExcellent PerformanceLevel = "excellent"
	PerformanceLevelGood      PerformanceLevel = "good"
	PerformanceLevelAverage   PerformanceLevel = "average"
	PerformanceLevelPoor      PerformanceLevel = "poor"
)

func (r SupplierRating) GetPerformanceLevel() PerformanceLevel {
	switch {
	case r.OverallScore >= 90:
		return PerformanceLevelExcellent
	case r.OverallScore >= 75:
		return PerformanceLevelGood
	case r.OverallScore >= 50:
		return PerformanceLevelAverage
	default:
		return PerformanceLevelPoor
	}
}

func (r SupplierRating) IsReliable() bool {
	return r.OverallScore >= 70 && r.RatingCount >= 5
}

func (p PerformanceLevel) String() string {
	return string(p)
}
