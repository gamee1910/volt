package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/restapi/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/restapi/handler/response"
)

type electricityService struct {
	electricityRepository repository.ElectricityRepository
	evnClient             service.EVNClient
}

func NewElectricityService(
	electricityRepository repository.ElectricityRepository,
	evnClient service.EVNClient,
) service.ElectricityService {
	return &electricityService{
		electricityRepository: electricityRepository,
		evnClient:             evnClient,
	}
}
func (s *electricityService) LoginEVN(ctx context.Context, username string, password string) error {
	return s.evnClient.Login(ctx, username, password)
}

func (s *electricityService) FetchAndSyncMonthlyUsage(
	ctx context.Context, req request.DailyPowerUsageRequest,
) (*response.MonthlyElectricityResponse, error) {
	resp, err := s.evnClient.GetDailyPowerUsageData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch EVN data: %w", err)
	}

	var totalKWh float64
	var consumptions []*entity.ElectricityConsumption

	for _, item := range resp.Data.DailyOutputs {
		readingDate := parseDate(item.MeasurementTimestamp, item.Date, item.FullDate)

		kwh, _ := strconv.ParseFloat(item.TotalOutput, 64)
		if kwh == 0 {
			kwh = item.TotalIndex
		}

		consumption := &entity.ElectricityConsumption{
			MeasurementDate: readingDate,
			ConsumptionKWh:  kwh,
		}

		if err = s.electricityRepository.Insert(ctx, consumption); err != nil {
			return nil, fmt.Errorf("failed to save consumption for date %s: %w", item.Date, err)
		}

		totalKWh += kwh
		consumptions = append(consumptions, consumption)
	}

	totalAmount := calculateElectricityBill(totalKWh)

	return &response.MonthlyElectricityResponse{
		TotalKWh:    totalKWh,
		TotalAmount: totalAmount,
		Data:        consumptions,
	}, nil
}

func (s *electricityService) GetAll(ctx context.Context) (*response.MonthlyElectricityResponse, error) {
	resp, err := s.electricityRepository.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	var totalKWh float64
	for _, v := range resp {
		totalKWh += v.ConsumptionKWh
	}
	totalAmount := calculateElectricityBill(totalKWh)

	return &response.MonthlyElectricityResponse{
		TotalKWh:    totalKWh,
		TotalAmount: totalAmount,
		Data:        resp,
	}, nil
}

func calculateElectricityBill(totalKWh float64) float64 {
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

func parseDate(timestampStr, dateStr, fullDateStr string) time.Time {
	layouts := []string{
		"02/01/2006 15:04:05",
		"02/01/2006",
		"2006-01-02 15:04:05",
		"2006-01-02",
	}

	for _, input := range []string{timestampStr, dateStr, fullDateStr} {
		if input == "" {
			continue
		}
		for _, layout := range layouts {
			if t, err := time.Parse(layout, input); err == nil {
				return t
			}
		}
	}
	return time.Time{}
}
