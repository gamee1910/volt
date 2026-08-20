package entity

import "time"

type ElectricityConsumption struct {
	ID             int64
	ReadingDate    time.Time
	ConsumptionKWh float64
	CreatedAt      time.Time
}
