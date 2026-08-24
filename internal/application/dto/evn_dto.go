package dto

type DailyPowerUsageRequest struct {
	Token        string
	CustomerCode string
	FromDate     string
	ToDate       string
}

type DailyPowerUsageResponse struct {
	State string
	Alert string
	Data  DailyPowerUsageData
}

type DailyPowerUsageData struct {
	NumberOfDays int
	Title        string
	DailyOutputs []DailyPowerUsage
}

type DailyPowerUsage struct {
	Date                 string
	FullDate             string
	OffPeakIndex         float64
	StandardIndex        float64
	PeakIndex            float64
	TotalIndex           float64
	OffPeakOutput        string
	StandardOutput       string
	PeakOutput           string
	TotalOutput          string
	MultiplicationFactor float64
	MeasurementTimestamp string
	IsBilled             int
}
