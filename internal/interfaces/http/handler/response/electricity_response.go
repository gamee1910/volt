package response

import "github.com/gamee1910/volt/internal/domain/entity"

type MonthlyElectricityResponse struct {
	TotalKWh    float64                          `json:"total_kwh"`
	TotalAmount float64                          `json:"total_amount"`
	Data        []*entity.ElectricityConsumption `json:"data"`
}
