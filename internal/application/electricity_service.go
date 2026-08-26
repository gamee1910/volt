package application

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/internal/domain/service"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/api/handler/response"
	"github.com/gamee1910/volt/pkg/evnhcmc"
	"github.com/gamee1910/volt/pkg/logger"
	"github.com/gamee1910/volt/pkg/utils"
)

type electricityService struct {
	electricityRepository repository.ElectricityRepository
	evnClient             evnhcmc.EVNClient
	log                   *logger.Logger
}

func NewElectricityService(
	electricityRepository repository.ElectricityRepository,
	evnClient evnhcmc.EVNClient,
	log *logger.Logger,
) service.ElectricityService {
	return &electricityService{
		electricityRepository: electricityRepository,
		evnClient:             evnClient,
		log:                   log,
	}
}

func (s *electricityService) DailyPowerUsage(
	ctx context.Context, req request.GetUsageRequest,
) error {
	resp, err := s.evnClient.GetDailyPowerUsageData(
		ctx,
		evnhcmc.DailyPowerUsageRequest{
			FromDate: req.FromDate,
			ToDate:   req.ToDate,
		})
	if err != nil {
		return fmt.Errorf("failed to fetch EVN data: %w", err)
	}

	for _, item := range resp.Data.DailyOutputs {
		kwh, err := strconv.ParseFloat(item.TotalOutput, 64)
		if err != nil {
			return fmt.Errorf(
				"invalid total output %q for date %s: %w",
				item.TotalOutput,
				item.Date,
				err,
			)
		}

		if kwh == 0 {
			kwh = item.TotalIndex
		}

		location, err := utils.LoadVietnamTimezone()
		if err != nil {
			return err
		}

		parseEVNDate, err := utils.ParseEVNDate(item.FullDate, location)
		if err != nil {
			return err
		}

		consumption := &entity.ElectricityConsumption{
			MeasurementDate: parseEVNDate,
			ConsumptionKWh:  kwh,
		}
		if err := s.electricityRepository.Upsert(ctx, consumption); err != nil {
			return fmt.Errorf(
				"failed to save consumption for date %s: %w",
				item.Date,
				err,
			)
		}
	}

	return nil
}

func (s *electricityService) GetAll(
	ctx context.Context,
) (*response.ElectricityResponse, error) {
	resp, err := s.electricityRepository.FetchAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch: %w", err)
	}

	var totalKWh float64
	var responseEntities []*response.ElectricityConsumptionResponse
	for _, v := range resp {
		var responseEntity = &response.ElectricityConsumptionResponse{
			MeasurementDate: v.MeasurementDate,
			ConsumptionKWh:  v.ConsumptionKWh,
			TotalAmount:     utils.CalculateElectricityBill(v.ConsumptionKWh),
		}

		responseEntities = append(responseEntities, responseEntity)
		totalKWh += v.ConsumptionKWh
	}
	totalAmount := utils.CalculateElectricityBill(totalKWh)

	return &response.ElectricityResponse{
		TotalKWh:    totalKWh,
		TotalAmount: totalAmount,
		Data:        responseEntities,
	}, nil
}

func (s *electricityService) GetYesterDayUsage(
	ctx context.Context,
) (*response.ElectricityConsumptionResponse, error) {
	loc, err := utils.LoadVietnamTimezone()
	if err != nil {
		return nil, err
	}

	now := time.Now().In(loc)
	yesterday := utils.GetYesterday(now, loc)
	firstDayOfMonth := utils.GetFirstDayOfMonth(now, loc)

	consumption, err := s.electricityRepository.FetchByDate(ctx, yesterday)
	if err != nil {
		return nil, fmt.Errorf("failed to get yesterday usage: %w", err)
	}

	totalKWh, err := s.electricityRepository.CalculateTotalConsumptionFromDateToDate(
		ctx,
		firstDayOfMonth,
		yesterday,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get monthly total usage: %w", err)
	}

	totalAmount := utils.CalculateElectricityBill(totalKWh)

	return &response.ElectricityConsumptionResponse{
		MeasurementDate: consumption.MeasurementDate,
		ConsumptionKWh:  consumption.ConsumptionKWh,
		TotalAmount:     totalAmount,
	}, nil
}
