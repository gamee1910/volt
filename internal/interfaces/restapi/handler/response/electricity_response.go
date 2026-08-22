package response

import (
	"encoding/json"
	"fmt"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const DateFormat = "02/01/2006"

type ElectricityResponse struct {
	TotalKWh    float64                           `json:"total_kwh"`
	TotalAmount float64                           `json:"total_amount"`
	Data        []*ElectricityConsumptionResponse `json:"data"`
}

func (r *ElectricityResponse) MarshalJSON() ([]byte, error) {
	type Alias ElectricityResponse
	return json.Marshal(&struct {
		TotalKWh    string `json:"total_kwh"`
		TotalAmount string `json:"total_amount"`
		*Alias
	}{
		TotalKWh:    fmt.Sprintf("%.2f KWh", r.TotalKWh),
		TotalAmount: formatVND(r.TotalAmount),
		Alias:       (*Alias)(r),
	})
}

type ElectricityConsumptionResponse struct {
	MeasurementDate time.Time `json:"measurement_date"`
	ConsumptionKWh  float64   `json:"consumption_kwh"`
	TotalAmount     float64   `json:"total_amount"`
}

func (r *ElectricityConsumptionResponse) MarshalJSON() ([]byte, error) {
	type Alias ElectricityConsumptionResponse
	return json.Marshal(&struct {
		MeasurementDate string `json:"measurement_date"`
		ConsumptionKWh  string `json:"consumption_kwh"`
		TotalAmount     string `json:"total_amount"`
		*Alias
	}{
		MeasurementDate: r.MeasurementDate.Format(DateFormat),
		ConsumptionKWh:  fmt.Sprintf("%.2f KWh", r.ConsumptionKWh),
		TotalAmount:     formatVND(r.TotalAmount),
		Alias:           (*Alias)(r),
	})
}

func formatVND(amount float64) string {
	p := message.NewPrinter(language.Vietnamese)
	return p.Sprintf("%.0f VND", amount)
}
