package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/http/handler/response"
	"github.com/gamee1910/volt/pkg/evnhcm"
)

type electricityService struct {
	electricityRepository repository.ElectricityRepository
	evnClient             *evnhcm.EVNClient
}

func NewElectricityService(
	electricityRepository repository.ElectricityRepository,
	evnClient *evnhcm.EVNClient,
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
	ctx context.Context,
	req evnhcm.DailyPowerUsageRequest,
) (*response.MonthlyElectricityResponse, error) {
	resp, err := s.evnClient.GetDailyPowerUsageData(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch EVN data: %w", err)
	}

	var totalKWh float64
	var consumptions []*entity.ElectricityConsumption

	for _, item := range resp.Data.DailyOutputs {
		readingDate, err := time.Parse("02/01/2006", item.Date)
		if err != nil {
			readingDate, _ = time.Parse("2006-01-02", item.FullDate)
		}

		kwh, _ := strconv.ParseFloat(item.TotalOutput, 64)
		if kwh == 0 {
			kwh = item.TotalIndex
		}

		consumption := &entity.ElectricityConsumption{
			ReadingDate:    readingDate,
			ConsumptionKWh: kwh,
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
