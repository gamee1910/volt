package types

import "time"

type GetUsageParam struct {
	FromDate string
	ToDate   string
}

type ElectricityConsumption struct {
	MeasurementDate time.Time
	ConsumptionKWh  float64
	TotalAmount     float64
}

type ElectricitySummary struct {
	TotalKWh    float64
	TotalAmount float64
	Data        []*ElectricityConsumption
}
