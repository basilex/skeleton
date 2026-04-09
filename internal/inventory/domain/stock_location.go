package domain

import (
	"errors"
	"fmt"
)

type StockLocation struct {
	warehouseID WarehouseID
	zone        string
	aisle       string
	bin         string
	description string
}

func NewStockLocation(warehouseID WarehouseID, zone, aisle, bin string) (StockLocation, error) {
	if zone == "" {
		return StockLocation{}, errors.New("zone is required")
	}
	if aisle == "" {
		return StockLocation{}, errors.New("aisle is required")
	}

	return StockLocation{
		warehouseID: warehouseID,
		zone:        zone,
		aisle:       aisle,
		bin:         bin,
	}, nil
}

func (l StockLocation) GetWarehouseID() WarehouseID { return l.warehouseID }
func (l StockLocation) GetZone() string             { return l.zone }
func (l StockLocation) GetAisle() string            { return l.aisle }
func (l StockLocation) GetBin() string              { return l.bin }
func (l StockLocation) GetDescription() string      { return l.description }

func (l StockLocation) SetDescription(desc string) StockLocation {
	l.description = desc
	return l
}

func (l StockLocation) IsZero() bool {
	return l.zone == "" && l.aisle == ""
}

func (l StockLocation) String() string {
	if l.bin != "" {
		return fmt.Sprintf("%s-%s-%s-%s", l.warehouseID, l.zone, l.aisle, l.bin)
	}
	return fmt.Sprintf("%s-%s-%s", l.warehouseID, l.zone, l.aisle)
}

func (l StockLocation) Equals(other StockLocation) bool {
	return l.warehouseID == other.warehouseID &&
		l.zone == other.zone &&
		l.aisle == other.aisle &&
		l.bin == other.bin
}

func (l StockLocation) ShortCode() string {
	return fmt.Sprintf("%s/%s/%s", l.zone, l.aisle, l.bin)
}

func (l StockLocation) FullCode() string {
	return fmt.Sprintf("%s-%s", l.warehouseID, l.ShortCode())
}

func WarehouseStockLocation(warehouseID WarehouseID) StockLocation {
	return StockLocation{
		warehouseID: warehouseID,
		zone:        "DEFAULT",
		aisle:       "001",
		bin:         "A001",
	}
}
