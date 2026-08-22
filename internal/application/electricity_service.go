package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/ports"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/response"
)

const (
	vietnamTimeZone = "Asia/Ho_Chi_Minh"
	startOfDay      = 0
	endOfMonth      = 1
)

type electricityService struct {
	electricityRepository ports.ElectricityRepository
	evnClient             ports.EVNClient
}

func NewElectricityService(
	electricityRepository ports.ElectricityRepository,
	evnClient ports.EVNClient,
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
) error {
	resp, err := s.evnClient.GetDailyPowerUsageData(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to fetch EVN data: %w", err)
	}

	var totalKWh float64
	var consumptions []*entity.ElectricityConsumption

	for _, item := range resp.Data.DailyOutputs {
		readingDate := s.parseDate(item.MeasurementTimestamp, item.Date, item.FullDate)

		kwh, _ := strconv.ParseFloat(item.TotalOutput, 64)
		if kwh == 0 {
			kwh = item.TotalIndex
		}

		consumption := &entity.ElectricityConsumption{
			MeasurementDate: readingDate,
			ConsumptionKWh:  kwh,
		}

		if err = s.electricityRepository.Upsert(ctx, consumption); err != nil {
			return fmt.Errorf("failed to save consumption for date %s: %w", item.Date, err)
		}

		totalKWh += kwh
		consumptions = append(consumptions, consumption)
	}

	return nil
}

func (s *electricityService) GetAll(ctx context.Context) (*response.ElectricityResponse, error) {
	resp, err := s.electricityRepository.GetAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	var totalKWh float64
	var responseEntities []*response.ElectricityConsumptionResponse
	for _, v := range resp {
		var responseEntity = &response.ElectricityConsumptionResponse{
			MeasurementDate: v.MeasurementDate,
			ConsumptionKWh:  v.ConsumptionKWh,
			TotalAmount:     s.calculateElectricityBill(v.ConsumptionKWh),
		}

		responseEntities = append(responseEntities, responseEntity)
		totalKWh += v.ConsumptionKWh
	}
	totalAmount := s.calculateElectricityBill(totalKWh)

	return &response.ElectricityResponse{
		TotalKWh:    totalKWh,
		TotalAmount: totalAmount,
		Data:        responseEntities,
	}, nil
}

func (s *electricityService) GetYesterDayUsage(ctx context.Context) (*response.ElectricityConsumptionResponse, error) {
	loc, err := s.loadVietnamTimezone()
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)
	yesterday := s.getYesterday(now, loc)
	firstDayOfMonth := s.getFirstDayOfMonth(now, loc)

	consumption, err := s.electricityRepository.GetByDate(ctx, yesterday)
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday usage: %w", err)
	}

	totalKWh, err := s.electricityRepository.GetTotalConsumption(
		ctx,
		firstDayOfMonth,
		yesterday,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly total usage: %w", err)
	}

	totalAmount := s.calculateElectricityBill(totalKWh)

	return &response.ElectricityConsumptionResponse{
		MeasurementDate: consumption.MeasurementDate,
		ConsumptionKWh:  consumption.ConsumptionKWh,
		TotalAmount:     totalAmount,
	}, nil
}

func (s *electricityService) loadVietnamTimezone() (*time.Location, error) {
	loc, err := time.LoadLocation(vietnamTimeZone)
	if err != nil {
		return nil, fmt.Errorf("load Vietnam timezone: %w", err)
	}
	return loc, nil
}

func (s *electricityService) getYesterday(now time.Time, loc *time.Location) time.Time {
	return time.Date(
		now.Year(),
		now.Month(),
		now.Day()-1,
		startOfDay, startOfDay, startOfDay, 0,
		loc,
	)
}

func (s *electricityService) getFirstDayOfMonth(now time.Time, loc *time.Location) time.Time {
	return time.Date(
		now.Year(),
		now.Month(),
		endOfMonth,
		startOfDay, startOfDay, startOfDay, 0,
		loc,
	)
}

func (s *electricityService) calculateElectricityBill(totalKWh float64) float64 {
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

func (s *electricityService) parseDate(timestampStr, dateStr, fullDateStr string) time.Time {
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
