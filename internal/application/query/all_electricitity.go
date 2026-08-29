package query

import (
	"context"
	"fmt"

	"github.com/gamee1910/volt/internal/application/types"
	"github.com/gamee1910/volt/internal/common/decorator"
	"github.com/gamee1910/volt/internal/common/utils"
	"github.com/gamee1910/volt/internal/domain/repository"
	"github.com/gamee1910/volt/pkg/logger"
)

type AllElectricityQuery struct{}

type AllElectricityHandler decorator.QueryHandler[
	AllElectricityQuery,
	types.ElectricitySummary,
]

type allElectricityHandler struct {
	repo repository.ElectricityRepository
}

func NewAllElectricityHandler(
	repo repository.ElectricityRepository, logger *logger.Logger, metricClient decorator.MetricClient,
) AllElectricityHandler {
	return decorator.ApplyQueryDecorators(
		allElectricityHandler{repo: repo}, logger, metricClient,
	)
}

func (h allElectricityHandler) Handle(
	ctx context.Context,
	_ AllElectricityQuery,
) (electricity types.ElectricitySummary, err error) {
	resp, err := h.repo.FetchAll(ctx)
	if err != nil {
		return types.ElectricitySummary{}, fmt.Errorf("failed to fetch all electricity: %w", err)
	}

	var (
		totalKWh         float64
		responseEntities []*types.ElectricityConsumption
	)

	for _, v := range resp {
		var responseEntity = &types.ElectricityConsumption{
			MeasurementDate: v.MeasurementDate,
			ConsumptionKWh:  v.ConsumptionKWh,
			TotalAmount:     utils.CalculateElectricityBill(v.ConsumptionKWh),
		}

		responseEntities = append(responseEntities, responseEntity)
		totalKWh += v.ConsumptionKWh
	}

	return types.ElectricitySummary{
		TotalKWh:    totalKWh,
		TotalAmount: utils.CalculateElectricityBill(totalKWh),
		Data:        responseEntities,
	}, nil
}
