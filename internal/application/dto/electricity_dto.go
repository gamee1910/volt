package dto

import "time"

type GetUsageParam struct {
	FromDate string
	ToDate   string
}

type ElectricityConsumptionDTO struct {
	MeasurementDate time.Time
	ConsumptionKWh  float64
	TotalAmount     float64
}

type ElectricitySummaryDTO struct {
	TotalKWh    float64
	TotalAmount float64
	Data        []*ElectricityConsumptionDTO
}
