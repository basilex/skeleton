package domain

import (
	"errors"
	"fmt"
	"time"
)

type LotID string

func NewLotID() LotID {
	return LotID(fmt.Sprintf("%d", time.Now().UnixNano()))
}

func (id LotID) String() string {
	return string(id)
}

type LotStatus string

const (
	LotStatusActive   LotStatus = "active"
	LotStatusExpired  LotStatus = "expired"
	LotStatusDepleted LotStatus = "depleted"
	LotStatusRecalled LotStatus = "recalled"
)

func (s LotStatus) String() string {
	return string(s)
}

type SerialNumberStatus string

const (
	SerialNumberAvailable SerialNumberStatus = "available"
	SerialNumberReserved  SerialNumberStatus = "reserved"
	SerialNumberSold      SerialNumberStatus = "sold"
	SerialNumberDefective SerialNumberStatus = "defective"
	SerialNumberReturned  SerialNumberStatus = "returned"
)

func (s SerialNumberStatus) String() string {
	return string(s)
}

type SerialNumber struct {
	number       string
	status       SerialNumberStatus
	reservedAt   *time.Time
	reservedBy   string
	soldAt       *time.Time
	defectReason string
}

func NewSerialNumber(number string) SerialNumber {
	return SerialNumber{
		number: number,
		status: SerialNumberAvailable,
	}
}

func (sn *SerialNumber) GetNumber() string             { return sn.number }
func (sn *SerialNumber) GetStatus() SerialNumberStatus { return sn.status }
func (sn *SerialNumber) GetReservedAt() *time.Time     { return sn.reservedAt }
func (sn *SerialNumber) GetReservedBy() string         { return sn.reservedBy }
func (sn *SerialNumber) GetSoldAt() *time.Time         { return sn.soldAt }

func (sn *SerialNumber) Reserve(reservedBy string) error {
	if sn.status != SerialNumberAvailable {
		return fmt.Errorf("serial number %s is not available", sn.number)
	}

	now := time.Now().UTC()
	sn.status = SerialNumberReserved
	sn.reservedAt = &now
	sn.reservedBy = reservedBy
	return nil
}

func (sn *SerialNumber) Release() error {
	if sn.status != SerialNumberReserved {
		return fmt.Errorf("serial number %s is not reserved", sn.number)
	}

	sn.status = SerialNumberAvailable
	sn.reservedAt = nil
	sn.reservedBy = ""
	return nil
}

func (sn *SerialNumber) MarkSold() error {
	if sn.status != SerialNumberReserved {
		return fmt.Errorf("serial number %s must be reserved before sale", sn.number)
	}

	now := time.Now().UTC()
	sn.status = SerialNumberSold
	sn.soldAt = &now
	return nil
}

func (sn *SerialNumber) MarkDefective(reason string) {
	sn.status = SerialNumberDefective
	sn.defectReason = reason
}

func (sn *SerialNumber) MarkReturned() error {
	if sn.status != SerialNumberSold {
		return fmt.Errorf("serial number %s is not sold", sn.number)
	}

	sn.status = SerialNumberReturned
	return nil
}

type Lot struct {
	id                LotID
	itemID            string
	warehouseID       WarehouseID
	lotNumber         string
	serialNumbers     []SerialNumber
	quantity          float64
	availableQty      float64
	manufacturingDate *time.Time
	expiryDate        *time.Time
	status            LotStatus
	location          *StockLocation
	createdAt         time.Time
	updatedAt         time.Time
	events            []DomainEvent
}

func NewLot(
	itemID string,
	warehouseID WarehouseID,
	lotNumber string,
	quantity float64,
	manufacturingDate *time.Time,
	expiryDate *time.Time,
) (*Lot, error) {
	if itemID == "" {
		return nil, errors.New("item ID is required")
	}
	if lotNumber == "" {
		return nil, errors.New("lot number is required")
	}
	if quantity <= 0 {
		return nil, errors.New("quantity must be positive")
	}

	now := time.Now().UTC()
	lot := &Lot{
		id:                NewLotID(),
		itemID:            itemID,
		warehouseID:       warehouseID,
		lotNumber:         lotNumber,
		serialNumbers:     make([]SerialNumber, 0),
		quantity:          quantity,
		availableQty:      quantity,
		manufacturingDate: manufacturingDate,
		expiryDate:        expiryDate,
		status:            LotStatusActive,
		createdAt:         now,
		updatedAt:         now,
		events:            make([]DomainEvent, 0),
	}

	lot.events = append(lot.events, LotCreated{
		LotID:      lot.id,
		ItemID:     itemID,
		LotNumber:  lotNumber,
		Quantity:   quantity,
		occurredAt: now,
	})

	return lot, nil
}

func (l *Lot) GetID() LotID                     { return l.id }
func (l *Lot) GetItemID() string                { return l.itemID }
func (l *Lot) GetWarehouseID() WarehouseID      { return l.warehouseID }
func (l *Lot) GetLotNumber() string             { return l.lotNumber }
func (l *Lot) GetSerialNumbers() []SerialNumber { return l.serialNumbers }
func (l *Lot) GetQuantity() float64             { return l.quantity }
func (l *Lot) GetAvailableQuantity() float64    { return l.availableQty }
func (l *Lot) GetManufacturingDate() *time.Time { return l.manufacturingDate }
func (l *Lot) GetExpiryDate() *time.Time        { return l.expiryDate }
func (l *Lot) GetStatus() LotStatus             { return l.status }
func (l *Lot) GetLocation() *StockLocation      { return l.location }
func (l *Lot) GetCreatedAt() time.Time          { return l.createdAt }
func (l *Lot) GetUpdatedAt() time.Time          { return l.updatedAt }

func (l *Lot) AddSerialNumber(number string) error {
	if l.status != LotStatusActive {
		return errors.New("cannot add serial number to non-active lot")
	}

	for _, sn := range l.serialNumbers {
		if sn.number == number {
			return fmt.Errorf("serial number %s already exists", number)
		}
	}

	l.serialNumbers = append(l.serialNumbers, NewSerialNumber(number))
	l.updatedAt = time.Now().UTC()
	return nil
}

func (l *Lot) ReserveSerialNumber(number string, reservedBy string) error {
	if l.status != LotStatusActive {
		return errors.New("lot is not active")
	}

	for i := range l.serialNumbers {
		if l.serialNumbers[i].number == number {
			if err := l.serialNumbers[i].Reserve(reservedBy); err != nil {
				return err
			}
			l.updatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("serial number %s not found", number)
}

func (l *Lot) ReleaseSerialNumber(number string) error {
	for i := range l.serialNumbers {
		if l.serialNumbers[i].number == number {
			if err := l.serialNumbers[i].Release(); err != nil {
				return err
			}
			l.updatedAt = time.Now().UTC()
			return nil
		}
	}

	return fmt.Errorf("serial number %s not found", number)
}

func (l *Lot) AdjustQuantity(delta float64) error {
	if l.status != LotStatusActive {
		return errors.New("cannot adjust quantity of non-active lot")
	}

	newQty := l.quantity + delta
	if newQty < 0 {
		return errors.New("quantity cannot be negative")
	}

	l.quantity = newQty
	l.availableQty = newQty
	l.updatedAt = time.Now().UTC()

	if l.quantity == 0 {
		l.status = LotStatusDepleted
	}

	l.events = append(l.events, LotQuantityAdjusted{
		LotID:      l.id,
		OldQty:     l.quantity - delta,
		NewQty:     l.quantity,
		occurredAt: l.updatedAt,
	})

	return nil
}

func (l *Lot) SetLocation(location StockLocation) {
	l.location = &location
	l.updatedAt = time.Now().UTC()
}

func (l *Lot) IsExpired() bool {
	if l.expiryDate == nil {
		return false
	}
	return time.Now().UTC().After(*l.expiryDate)
}

func (l *Lot) MarkExpired() {
	if l.status != LotStatusActive {
		return
	}
	l.status = LotStatusExpired
	l.updatedAt = time.Now().UTC()

	l.events = append(l.events, LotExpired{
		LotID:      l.id,
		occurredAt: l.updatedAt,
	})
}

func (l *Lot) Recall(reason string) {
	l.status = LotStatusRecalled
	l.updatedAt = time.Now().UTC()

	l.events = append(l.events, LotRecalled{
		LotID:      l.id,
		Reason:     reason,
		occurredAt: l.updatedAt,
	})
}

func (l *Lot) IsExpiredWithin(days int) bool {
	if l.expiryDate == nil {
		return false
	}

	expiryThreshold := time.Now().UTC().AddDate(0, 0, days)
	return l.expiryDate.Before(expiryThreshold) || l.expiryDate.Equal(expiryThreshold)
}

func (l *Lot) PullEvents() []DomainEvent {
	events := l.events
	l.events = make([]DomainEvent, 0)
	return events
}

func (l *Lot) String() string {
	return fmt.Sprintf("Lot{id=%s, lotNumber=%s, qty=%.2f, status=%s}",
		l.id, l.lotNumber, l.quantity, l.status)
}

func ReconstituteLot(
	id LotID,
	itemID string,
	warehouseID WarehouseID,
	lotNumber string,
	serialNumbers []SerialNumber,
	quantity float64,
	availableQty float64,
	manufacturingDate *time.Time,
	expiryDate *time.Time,
	status LotStatus,
	location *StockLocation,
	createdAt time.Time,
	updatedAt time.Time,
) *Lot {
	return &Lot{
		id:                id,
		itemID:            itemID,
		warehouseID:       warehouseID,
		lotNumber:         lotNumber,
		serialNumbers:     serialNumbers,
		quantity:          quantity,
		availableQty:      availableQty,
		manufacturingDate: manufacturingDate,
		expiryDate:        expiryDate,
		status:            status,
		location:          location,
		createdAt:         createdAt,
		updatedAt:         updatedAt,
		events:            make([]DomainEvent, 0),
	}
}
