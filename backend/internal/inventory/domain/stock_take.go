package domain

import (
	"errors"
	"fmt"
	"time"
)

type StockTakeID string

func NewStockTakeID() StockTakeID {
	return StockTakeID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id StockTakeID) String() string {
	return string(id)
}

type StockTakeStatus string

const (
	StockTakeStatusPending    StockTakeStatus = "pending"
	StockTakeStatusInProgress StockTakeStatus = "in_progress"
	StockTakeStatusCompleted  StockTakeStatus = "completed"
	StockTakeStatusCancelled  StockTakeStatus = "cancelled"
)

func (s StockTakeStatus) String() string {
	return string(s)
}

type StockTakeLine struct {
	itemID     string
	systemQty  float64
	countedQty float64
	variance   float64
	reason     string
	status     StockTakeLineStatus
	countedBy  string
	countedAt  *time.Time
}

type StockTakeLineStatus string

const (
	StockTakeLinePending  StockTakeLineStatus = "pending"
	StockTakeLineCounted  StockTakeLineStatus = "counted"
	StockTakeLineAdjusted StockTakeLineStatus = "adjusted"
	StockTakeLineVariance StockTakeLineStatus = "variance"
)

func NewStockTakeLine(systemQty float64) StockTakeLine {
	variance := 0.0 - systemQty
	return StockTakeLine{
		systemQty:  systemQty,
		countedQty: 0,
		variance:   variance,
		status:     StockTakeLinePending,
	}
}

func (l *StockTakeLine) GetItemID() string              { return l.itemID }
func (l *StockTakeLine) GetSystemQty() float64          { return l.systemQty }
func (l *StockTakeLine) GetCountedQty() float64         { return l.countedQty }
func (l *StockTakeLine) GetVariance() float64           { return l.variance }
func (l *StockTakeLine) GetReason() string              { return l.reason }
func (l *StockTakeLine) GetStatus() StockTakeLineStatus { return l.status }
func (l *StockTakeLine) GetCountedBy() string           { return l.countedBy }
func (l *StockTakeLine) GetCountedAt() *time.Time       { return l.countedAt }

func (l *StockTakeLine) Count(countedQty float64, countedBy string) {
	l.countedQty = countedQty
	l.variance = countedQty - l.systemQty
	l.countedBy = countedBy
	now := time.Now().UTC()
	l.countedAt = &now

	if l.variance != 0 {
		l.status = StockTakeLineVariance
	} else {
		l.status = StockTakeLineCounted
	}
}

func (l *StockTakeLine) SetReason(reason string) {
	l.reason = reason
}

func (l *StockTakeLine) Adjust() {
	l.status = StockTakeLineAdjusted
}

type StockTake struct {
	id          StockTakeID
	warehouseID WarehouseID
	reference   string
	lines       map[string]StockTakeLine
	status      StockTakeStatus
	startedAt   *time.Time
	startedBy   string
	completedAt *time.Time
	completedBy string
	createdAt   time.Time
	updatedAt   time.Time
	events      []DomainEvent
}

func NewStockTake(warehouseID WarehouseID, reference string, createdBy string) (*StockTake, error) {
	if reference == "" {
		return nil, errors.New("reference is required")
	}

	now := time.Now().UTC()
	stockTake := &StockTake{
		id:          NewStockTakeID(),
		warehouseID: warehouseID,
		reference:   reference,
		lines:       make(map[string]StockTakeLine),
		status:      StockTakeStatusPending,
		createdAt:   now,
		updatedAt:   now,
		events:      make([]DomainEvent, 0),
	}

	stockTake.events = append(stockTake.events, StockTakeCreated{
		StockTakeID: stockTake.id,
		WarehouseID: warehouseID,
		Reference:   reference,
		occurredAt:  now,
	})

	return stockTake, nil
}

func (s *StockTake) GetID() StockTakeID                 { return s.id }
func (s *StockTake) GetWarehouseID() WarehouseID        { return s.warehouseID }
func (s *StockTake) GetReference() string               { return s.reference }
func (s *StockTake) GetLines() map[string]StockTakeLine { return s.lines }
func (s *StockTake) GetStatus() StockTakeStatus         { return s.status }
func (s *StockTake) GetStartedAt() *time.Time           { return s.startedAt }
func (s *StockTake) GetStartedBy() string               { return s.startedBy }
func (s *StockTake) GetCompletedAt() *time.Time         { return s.completedAt }
func (s *StockTake) GetCompletedBy() string             { return s.completedBy }
func (s *StockTake) GetCreatedAt() time.Time            { return s.createdAt }
func (s *StockTake) GetUpdatedAt() time.Time            { return s.updatedAt }

func (s *StockTake) AddItem(itemID string, systemQty float64) error {
	if s.status != StockTakeStatusPending && s.status != StockTakeStatusInProgress {
		return errors.New("cannot add items to completed or cancelled stock take")
	}

	if _, exists := s.lines[itemID]; exists {
		return fmt.Errorf("item %s already added", itemID)
	}

	s.lines[itemID] = StockTakeLine{
		itemID:     itemID,
		systemQty:  systemQty,
		countedQty: 0,
		variance:   -systemQty,
		status:     StockTakeLinePending,
	}
	s.updatedAt = time.Now().UTC()

	return nil
}

func (s *StockTake) CountItem(itemID string, countedQty float64, countedBy string) error {
	if s.status == StockTakeStatusCompleted || s.status == StockTakeStatusCancelled {
		return errors.New("cannot count items in completed or cancelled stock take")
	}

	line, exists := s.lines[itemID]
	if !exists {
		return fmt.Errorf("item %s not found in stock take", itemID)
	}

	line.Count(countedQty, countedBy)
	s.lines[itemID] = line
	s.updatedAt = time.Now().UTC()

	s.events = append(s.events, StockTakeItemCounted{
		StockTakeID: s.id,
		ItemID:      itemID,
		SystemQty:   line.systemQty,
		CountedQty:  countedQty,
		Variance:    line.variance,
		occurredAt:  s.updatedAt,
	})

	return nil
}

func (s *StockTake) SetVarianceReason(itemID string, reason string) error {
	line, exists := s.lines[itemID]
	if !exists {
		return fmt.Errorf("item %s not found", itemID)
	}

	line.SetReason(reason)
	s.lines[itemID] = line
	s.updatedAt = time.Now().UTC()
	return nil
}

func (s *StockTake) Start(startedBy string) error {
	if s.status != StockTakeStatusPending {
		return errors.New("stock take already started")
	}

	now := time.Now().UTC()
	s.status = StockTakeStatusInProgress
	s.startedAt = &now
	s.startedBy = startedBy
	s.updatedAt = now

	s.events = append(s.events, StockTakeStarted{
		StockTakeID: s.id,
		StartedBy:   startedBy,
		occurredAt:  now,
	})

	return nil
}

func (s *StockTake) Complete(completedBy string) error {
	if s.status != StockTakeStatusInProgress {
		return errors.New("stock take must be in progress to complete")
	}

	pendingCount := 0
	for _, line := range s.lines {
		if line.status == StockTakeLinePending {
			pendingCount++
		}
	}

	if pendingCount > 0 {
		return fmt.Errorf("%d items have not been counted", pendingCount)
	}

	now := time.Now().UTC()
	s.status = StockTakeStatusCompleted
	s.completedAt = &now
	s.completedBy = completedBy
	s.updatedAt = now

	s.events = append(s.events, StockTakeCompleted{
		StockTakeID: s.id,
		CompletedBy: completedBy,
		occurredAt:  now,
	})

	return nil
}

func (s *StockTake) Cancel(reason string) error {
	if s.status == StockTakeStatusCompleted {
		return errors.New("cannot cancel completed stock take")
	}

	s.status = StockTakeStatusCancelled
	s.updatedAt = time.Now().UTC()

	s.events = append(s.events, StockTakeCancelled{
		StockTakeID: s.id,
		Reason:      reason,
		occurredAt:  s.updatedAt,
	})

	return nil
}

func (s *StockTake) GetVariances() map[string]float64 {
	variances := make(map[string]float64)
	for itemID, line := range s.lines {
		if line.variance != 0 {
			variances[itemID] = line.variance
		}
	}
	return variances
}

func (s *StockTake) HasVariance() bool {
	for _, line := range s.lines {
		if line.variance != 0 {
			return true
		}
	}
	return false
}

func (s *StockTake) GetVarianceCount() int {
	count := 0
	for _, line := range s.lines {
		if line.variance != 0 {
			count++
		}
	}
	return count
}

func (s *StockTake) GetTotalItems() int {
	return len(s.lines)
}

func (s *StockTake) GetCountedItems() int {
	count := 0
	for _, line := range s.lines {
		if line.status == StockTakeLineCounted || line.status == StockTakeLineVariance || line.status == StockTakeLineAdjusted {
			count++
		}
	}
	return count
}

func (s *StockTake) PullEvents() []DomainEvent {
	events := s.events
	s.events = make([]DomainEvent, 0)
	return events
}

func (s *StockTake) String() string {
	return fmt.Sprintf("StockTake{id=%s, warehouse=%s, status=%s, items=%d}",
		s.id, s.warehouseID, s.status, len(s.lines))
}

func ReconstituteStockTake(
	id StockTakeID,
	warehouseID WarehouseID,
	reference string,
	lines map[string]StockTakeLine,
	status StockTakeStatus,
	startedAt *time.Time,
	startedBy string,
	completedAt *time.Time,
	completedBy string,
	createdAt time.Time,
	updatedAt time.Time,
) *StockTake {
	return &StockTake{
		id:          id,
		warehouseID: warehouseID,
		reference:   reference,
		lines:       lines,
		status:      status,
		startedAt:   startedAt,
		startedBy:   startedBy,
		completedAt: completedAt,
		completedBy: completedBy,
		createdAt:   createdAt,
		updatedAt:   updatedAt,
		events:      make([]DomainEvent, 0),
	}
}
