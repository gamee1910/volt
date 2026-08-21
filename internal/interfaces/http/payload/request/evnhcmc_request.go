package request

type DailyPowerUsageRequest struct {
	Token        string
	CustomerCode string
	FromDate     string
	ToDate       string
}
