package request

type GetUsageRequest struct {
	FromDate string `json:"from_date"`
	ToDate   string `json:"to_date"`
}
