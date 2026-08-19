package entity

import "time"

type ElectricityConsumption struct {
	CustomerCode string
	Date         time.Time

	TD    float64
	BT    float64
	CD    float64
	Total float64

	PowerTD    string
	PowerBT    string
	PowerCD    string
	PowerTotal string

	HSN float64

	PGiaoBT    string
	PGiaoTD    string
	PGiaoCD    string
	TotalPGiao string
	TotalQGiao string

	IsChotHoaDon bool
}
