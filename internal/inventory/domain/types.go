package domain

type WarehouseStatus string

const (
	WarehouseStatusActive      WarehouseStatus = "active"
	WarehouseStatusInactive    WarehouseStatus = "inactive"
	WarehouseStatusMaintenance WarehouseStatus = "maintenance"
)

func (s WarehouseStatus) String() string {
	return string(s)
}

func (s WarehouseStatus) CanTransitionTo(newStatus WarehouseStatus) bool {
	switch s {
	case WarehouseStatusActive:
		return newStatus == WarehouseStatusInactive || newStatus == WarehouseStatusMaintenance
	case WarehouseStatusInactive:
		return newStatus == WarehouseStatusActive
	case WarehouseStatusMaintenance:
		return newStatus == WarehouseStatusActive || newStatus == WarehouseStatusInactive
	default:
		return false
	}
}

type MovementType string

const (
	MovementTypeReceipt    MovementType = "receipt"
	MovementTypeIssue      MovementType = "issue"
	MovementTypeTransfer   MovementType = "transfer"
	MovementTypeAdjustment MovementType = "adjustment"
	MovementTypeReturn     MovementType = "return"
)

func (t MovementType) String() string {
	return string(t)
}

func (t MovementType) IsInbound() bool {
	return t == MovementTypeReceipt || t == MovementTypeReturn || t == MovementTypeAdjustment
}

func (t MovementType) IsOutbound() bool {
	return t == MovementTypeIssue || t == MovementTypeTransfer
}

type ReservationStatus string

const (
	ReservationStatusActive    ReservationStatus = "active"
	ReservationStatusFulfilled ReservationStatus = "fulfilled"
	ReservationStatusCancelled ReservationStatus = "cancelled"
	ReservationStatusExpired   ReservationStatus = "expired"
)

func (s ReservationStatus) String() string {
	return string(s)
}

func (s ReservationStatus) IsActive() bool {
	return s == ReservationStatusActive
}
