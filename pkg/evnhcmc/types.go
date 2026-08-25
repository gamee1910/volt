package evnhcmc

type DailyPowerUsageRequest struct {
	Token        string
	CustomerCode string
	FromDate     string
	ToDate       string
}
type DailyPowerUsageResponse struct {
	State string              `json:"state"`
	Alert string              `json:"alert"`
	Data  DailyPowerUsageData `json:"data"`
}

type DailyPowerUsageData struct {
	NumberOfDays int               `json:"soNgay"`
	Title        string            `json:"tieude"`
	DailyOutputs []DailyPowerUsage `json:"sanluong_tungngay"`
}

type DailyPowerUsage struct {
	Date                 string  `json:"ngay"`
	FullDate             string  `json:"ngayFull"`
	OffPeakIndex         float64 `json:"TD"`
	StandardIndex        float64 `json:"BT"`
	PeakIndex            float64 `json:"CD"`
	TotalIndex           float64 `json:"Tong"`
	OffPeakOutput        string  `json:"sanluong_TD"`
	StandardOutput       string  `json:"sanluong_BT"`
	PeakOutput           string  `json:"sanluong_CD"`
	TotalOutput          string  `json:"sanluong_tong"`
	MultiplicationFactor float64 `json:"hsn"`
	MeasurementTimestamp string  `json:"thoidiemdo"`
	IsBilled             int     `json:"isChotHoaDon"`
}
