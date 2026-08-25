package utils

import (
	"fmt"
	"time"

	"golang.org/x/text/language"
	"golang.org/x/text/message"
)

const (
	vietnamTimeZone = "Asia/Ho_Chi_Minh"
	startOfDay      = 0
	firstDay        = 1
	evnDateFormat   = "02/01/2006"
)

func LoadVietnamTimezone() (*time.Location, error) {
	loc, err := time.LoadLocation(vietnamTimeZone)
	if err != nil {
		return nil, fmt.Errorf("load Vietnam timezone: %w", err)
	}
	return loc, nil
}

func GetYesterday(now time.Time, loc *time.Location) time.Time {
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day()-1,
		startOfDay, startOfDay, startOfDay, 0,
		loc,
	)
}
func GetFirstDayOfMonth(now time.Time, loc *time.Location) time.Time {
	return time.Date(
		now.Year(),
		now.Month(),
		firstDay,
		startOfDay,
		startOfDay,
		startOfDay,
		0,
		loc,
	)
}

func CalculateElectricityBill(totalKWh float64) float64 {
	tiers := []struct {
		limit float64
		price float64
	}{
		{50, 1893},
		{50, 1956},  // 51-100
		{100, 2271}, // 101-200
		{100, 2860}, // 201-300
		{100, 3197}, // 301-400
		{0, 3302},   // > 400
	}

	var totalAmount float64
	remainingKWh := totalKWh
	for _, tier := range tiers {
		if remainingKWh <= 0 {
			break
		}
		if tier.limit == 0 || remainingKWh <= tier.limit {
			totalAmount += remainingKWh * tier.price
			break
		}
		totalAmount += tier.limit * tier.price
		remainingKWh -= tier.limit
	}
	// Cộng thêm 8% thuế VAT
	totalAmountWithVAT := totalAmount * 1.08
	return totalAmountWithVAT
}

func FormatVND(amount float64) string {
	p := message.NewPrinter(language.Vietnamese)
	return p.Sprintf("%.0f VND", amount)
}

func FormatEVNDate(t time.Time) string {
	return t.Format(evnDateFormat)
}

func ParseEVNDate(value string, loc *time.Location) (time.Time, error) {
	return time.ParseInLocation(evnDateFormat, value, loc)
}
