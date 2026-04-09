package catalog

import (
	"errors"
	"fmt"
	"time"
)

type PricingRuleID string

func NewPricingRuleID() PricingRuleID {
	return PricingRuleID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id PricingRuleID) String() string {
	return string(id)
}

type PricingRuleType string

const (
	PricingRuleTypeVolumeDiscount PricingRuleType = "volume_discount"
	PricingRuleTypeCustomerGroup  PricingRuleType = "customer_group"
	PricingRuleTypeDateRange      PricingRuleType = "date_range"
	PricingRuleTypeBundle         PricingRuleType = "bundle"
)

func (t PricingRuleType) String() string {
	return string(t)
}

type PricingRuleStatus string

const (
	PricingRuleStatusActive   PricingRuleStatus = "active"
	PricingRuleStatusInactive PricingRuleStatus = "inactive"
	PricingRuleStatusExpired  PricingRuleStatus = "expired"
)

func (s PricingRuleStatus) String() string {
	return string(s)
}

type PricingRule struct {
	id              PricingRuleID
	name            string
	ruleType        PricingRuleType
	itemIDs         []ItemID
	categoryIDs     []CategoryID
	customerGroups  []string
	minQuantity     int
	maxQuantity     int
	discountPercent float64
	discountAmount  float64
	startDate       *time.Time
	endDate         *time.Time
	priority        int
	status          PricingRuleStatus
	createdAt       time.Time
	updatedAt       time.Time
	events          []DomainEvent
}

func NewPricingRule(
	name string,
	ruleType PricingRuleType,
	minQuantity int,
	discountPercent float64,
	discountAmount float64,
) (*PricingRule, error) {
	if name == "" {
		return nil, errors.New("rule name is required")
	}
	if discountPercent < 0 || discountPercent > 100 {
		return nil, errors.New("discount percent must be between 0 and 100")
	}
	if discountAmount < 0 {
		return nil, errors.New("discount amount cannot be negative")
	}
	if minQuantity < 1 {
		return nil, errors.New("minimum quantity must be at least 1")
	}

	now := time.Now().UTC()
	rule := &PricingRule{
		id:              NewPricingRuleID(),
		name:            name,
		ruleType:        ruleType,
		itemIDs:         make([]ItemID, 0),
		categoryIDs:     make([]CategoryID, 0),
		customerGroups:  make([]string, 0),
		minQuantity:     minQuantity,
		maxQuantity:     0,
		discountPercent: discountPercent,
		discountAmount:  discountAmount,
		priority:        0,
		status:          PricingRuleStatusActive,
		createdAt:       now,
		updatedAt:       now,
		events:          make([]DomainEvent, 0),
	}

	rule.events = append(rule.events, PricingRuleCreated{
		RuleID:      rule.id,
		Name:        rule.name,
		RuleType:    rule.ruleType,
		MinQuantity: rule.minQuantity,
		occurredAt:  now,
	})

	return rule, nil
}

func (r *PricingRule) GetID() PricingRuleID         { return r.id }
func (r *PricingRule) GetName() string              { return r.name }
func (r *PricingRule) GetRuleType() PricingRuleType { return r.ruleType }
func (r *PricingRule) GetItemIDs() []ItemID         { return r.itemIDs }
func (r *PricingRule) GetCategoryIDs() []CategoryID { return r.categoryIDs }
func (r *PricingRule) GetCustomerGroups() []string  { return r.customerGroups }
func (r *PricingRule) GetMinQuantity() int          { return r.minQuantity }
func (r *PricingRule) GetMaxQuantity() int          { return r.maxQuantity }
func (r *PricingRule) GetDiscountPercent() float64  { return r.discountPercent }
func (r *PricingRule) GetDiscountAmount() float64   { return r.discountAmount }
func (r *PricingRule) GetStartDate() *time.Time     { return r.startDate }
func (r *PricingRule) GetEndDate() *time.Time       { return r.endDate }
func (r *PricingRule) GetPriority() int             { return r.priority }
func (r *PricingRule) GetStatus() PricingRuleStatus { return r.status }
func (r *PricingRule) GetCreatedAt() time.Time      { return r.createdAt }
func (r *PricingRule) GetUpdatedAt() time.Time      { return r.updatedAt }

func (r *PricingRule) AddItem(itemID ItemID) error {
	if r.status != PricingRuleStatusActive {
		return errors.New("cannot modify inactive or expired rule")
	}

	for _, id := range r.itemIDs {
		if id == itemID {
			return nil
		}
	}

	r.itemIDs = append(r.itemIDs, itemID)
	r.updatedAt = time.Now().UTC()
	return nil
}

func (r *PricingRule) AddCategory(categoryID CategoryID) error {
	if r.status != PricingRuleStatusActive {
		return errors.New("cannot modify inactive or expired rule")
	}

	for _, id := range r.categoryIDs {
		if id == categoryID {
			return nil
		}
	}

	r.categoryIDs = append(r.categoryIDs, categoryID)
	r.updatedAt = time.Now().UTC()
	return nil
}

func (r *PricingRule) AddCustomerGroup(group string) error {
	if r.status != PricingRuleStatusActive {
		return errors.New("cannot modify inactive or expired rule")
	}

	for _, g := range r.customerGroups {
		if g == group {
			return nil
		}
	}

	r.customerGroups = append(r.customerGroups, group)
	r.updatedAt = time.Now().UTC()
	return nil
}

func (r *PricingRule) SetDateRange(start, end *time.Time) error {
	if start != nil && end != nil && start.After(*end) {
		return errors.New("start date must be before end date")
	}

	r.startDate = start
	r.endDate = end
	r.updatedAt = time.Now().UTC()
	return nil
}

func (r *PricingRule) SetQuantityRange(min, max int) error {
	if min < 1 {
		return errors.New("minimum quantity must be at least 1")
	}
	if max > 0 && min > max {
		return errors.New("minimum cannot be greater than maximum")
	}

	r.minQuantity = min
	r.maxQuantity = max
	r.updatedAt = time.Now().UTC()
	return nil
}

func (r *PricingRule) SetPriority(priority int) {
	r.priority = priority
	r.updatedAt = time.Now().UTC()
}

func (r *PricingRule) UpdateDiscount(percent, amount float64) error {
	if percent < 0 || percent > 100 {
		return errors.New("discount percent must be between 0 and 100")
	}
	if amount < 0 {
		return errors.New("discount amount cannot be negative")
	}

	r.discountPercent = percent
	r.discountAmount = amount
	r.updatedAt = time.Now().UTC()
	return nil
}

func (r *PricingRule) Activate() {
	if r.status == PricingRuleStatusActive {
		return
	}

	r.status = PricingRuleStatusActive
	r.updatedAt = time.Now().UTC()
}

func (r *PricingRule) Deactivate() {
	if r.status == PricingRuleStatusInactive {
		return
	}

	r.status = PricingRuleStatusInactive
	r.updatedAt = time.Now().UTC()
}

func (r *PricingRule) IsApplicable(itemID ItemID, quantity int, customerGroup string, now time.Time) bool {
	if r.status != PricingRuleStatusActive {
		return false
	}

	if r.startDate != nil && now.Before(*r.startDate) {
		return false
	}

	if r.endDate != nil && now.After(*r.endDate) {
		return false
	}

	if quantity < r.minQuantity {
		return false
	}

	if r.maxQuantity > 0 && quantity > r.maxQuantity {
		return false
	}

	itemMatch := len(r.itemIDs) == 0
	for _, id := range r.itemIDs {
		if id == itemID {
			itemMatch = true
			break
		}
	}
	if !itemMatch {
		return false
	}

	if len(r.customerGroups) > 0 && customerGroup != "" {
		groupMatch := false
		for _, g := range r.customerGroups {
			if g == customerGroup {
				groupMatch = true
				break
			}
		}
		if !groupMatch {
			return false
		}
	}

	return true
}

func (r *PricingRule) CalculateDiscount(basePrice float64, quantity int) float64 {
	if !r.IsApplicable(ItemID{}, quantity, "", time.Now()) {
		return 0
	}

	var discount float64
	if r.discountPercent > 0 {
		discount = basePrice * (r.discountPercent / 100)
	}
	if r.discountAmount > 0 {
		discount += r.discountAmount
	}

	return discount
}

func (r *PricingRule) PullEvents() []DomainEvent {
	events := r.events
	r.events = make([]DomainEvent, 0)
	return events
}

func (r *PricingRule) String() string {
	return fmt.Sprintf("PricingRule{id=%s, name=%s, type=%s, status=%s}",
		r.id, r.name, r.ruleType, r.status)
}

func ReconstitutePricingRule(
	id PricingRuleID,
	name string,
	ruleType PricingRuleType,
	itemIDs []ItemID,
	categoryIDs []CategoryID,
	customerGroups []string,
	minQuantity, maxQuantity int,
	discountPercent, discountAmount float64,
	startDate, endDate *time.Time,
	priority int,
	status PricingRuleStatus,
	createdAt, updatedAt time.Time,
) *PricingRule {
	return &PricingRule{
		id:              id,
		name:            name,
		ruleType:        ruleType,
		itemIDs:         itemIDs,
		categoryIDs:     categoryIDs,
		customerGroups:  customerGroups,
		minQuantity:     minQuantity,
		maxQuantity:     maxQuantity,
		discountPercent: discountPercent,
		discountAmount:  discountAmount,
		startDate:       startDate,
		endDate:         endDate,
		priority:        priority,
		status:          status,
		createdAt:       createdAt,
		updatedAt:       updatedAt,
		events:          make([]DomainEvent, 0),
	}
}
