package application

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gamee1910/volt/internal/domain/entity"
	"github.com/gamee1910/volt/internal/interfaces/http/transport/request"
	"github.com/gamee1910/volt/internal/interfaces/http/transport/response"
)

type mockEVNClient struct {
	response *response.DailyPowerUsageResponse
	err      error
}

func (m *mockEVNClient) Login(
	ctx context.Context, username string, password string,
) error {
	return nil
}

func (m *mockEVNClient) GetDailyPowerUsageData(
	ctx context.Context, req request.DailyPowerUsageRequest,
) (*response.DailyPowerUsageResponse, error) {
	return m.response, m.err
}

type mockElectricityRepository struct {
	consumptions []*entity.ElectricityConsumption
	err          error
}

func (m *mockElectricityRepository) Insert(
	ctx context.Context, consumption *entity.ElectricityConsumption,
) error {
	if m.err != nil {
		return m.err
	}

	m.consumptions = append(
		m.consumptions,
		consumption,
	)

	return nil
}

func (m *mockElectricityRepository) FindByDate(
	ctx context.Context, date time.Time,
) (*entity.ElectricityConsumption, error) {
	if m.err != nil {
		return nil, m.err
	}
	for _, c := range m.consumptions {
		if c.MeasurementDate.Equal(date) {
			return c, nil
		}
	}
	return nil, nil
}

func (m *mockElectricityRepository) FindAll(
	ctx context.Context,
) ([]*entity.ElectricityConsumption, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.consumptions, nil
}

func newMockEVNResponse(
	outputs ...response.DailyPowerUsage,
) *response.DailyPowerUsageResponse {
	return &response.DailyPowerUsageResponse{
		Data: response.DailyPowerUsageData{
			DailyOutputs: outputs,
		},
	}
}

func TestElectricityService_FetchAndSyncMonthlyUsage_Success(t *testing.T) {
	ctx := context.Background()

	evnMock := &mockEVNClient{
		response: newMockEVNResponse(
			response.DailyPowerUsage{
				MeasurementTimestamp: "01/08/2026 00:00:00",
				Date:                 "01/08/2026",
				FullDate:             "2026-08-01",
				TotalOutput:          "10",
				TotalIndex:           0,
			},
			response.DailyPowerUsage{
				Date:        "02/08/2026",
				FullDate:    "2026-08-02",
				TotalOutput: "20",
				TotalIndex:  0,
			},
			response.DailyPowerUsage{
				Date:        "invalid-date",
				FullDate:    "2026-08-03",
				TotalOutput: "0",
				TotalIndex:  30,
			},
		),
	}

	repositoryMock := &mockElectricityRepository{}

	svc := NewElectricityService(repositoryMock, evnMock)

	req := request.DailyPowerUsageRequest{}

	result, err := svc.FetchAndSyncMonthlyUsage(ctx, req)

	require.NoError(t, err)
	require.NotNil(t, result)

	assert.Equal(t, 60.0, result.TotalKWh)

	require.Len(t, result.Data, 3)
	require.Len(t, repositoryMock.consumptions, 3)

	first := repositoryMock.consumptions[0]

	assert.Equal(
		t,
		time.Date(2026, time.August, 1, 0, 0, 0, 0, time.UTC),
		first.MeasurementDate,
	)
	assert.Equal(t, 10.0, first.ConsumptionKWh)

	second := repositoryMock.consumptions[1]

	assert.Equal(
		t,
		time.Date(2026, time.August, 2, 0, 0, 0, 0, time.UTC),
		second.MeasurementDate,
	)
	assert.Equal(t, 20.0, second.ConsumptionKWh)

	third := repositoryMock.consumptions[2]

	assert.Equal(
		t,
		time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC),
		third.MeasurementDate,
	)
	assert.Equal(t, 30.0, third.ConsumptionKWh)

	assert.InDelta(t, 123346.8, result.TotalAmount, 0.001)
}

func TestElectricityService_FetchAndSyncMonthlyUsage_EVNError(t *testing.T) {
	ctx := context.Background()

	evnMock := &mockEVNClient{
		err: errors.New("network connection error"),
	}

	repositoryMock := &mockElectricityRepository{}

	svc := NewElectricityService(repositoryMock, evnMock)

	req := request.DailyPowerUsageRequest{}

	result, err := svc.FetchAndSyncMonthlyUsage(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to fetch EVN data")
}

func TestElectricityService_FetchAndSyncMonthlyUsage_RepositoryError(t *testing.T) {
	ctx := context.Background()

	evnMock := &mockEVNClient{
		response: newMockEVNResponse(
			response.DailyPowerUsage{
				Date:        "01/08/2026",
				FullDate:    "2026-08-01",
				TotalOutput: "10",
			},
		),
	}

	repositoryMock := &mockElectricityRepository{
		err: errors.New("database connection failed"),
	}

	svc := NewElectricityService(repositoryMock, evnMock)

	req := request.DailyPowerUsageRequest{}

	result, err := svc.FetchAndSyncMonthlyUsage(ctx, req)

	require.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to save consumption for date")
}

func TestCalculateElectricityBill(t *testing.T) {
	tests := []struct {
		name     string
		totalKWh float64
		expected float64
	}{
		{
			name:     "0 kWh",
			totalKWh: 0,
			expected: 0,
		},
		{
			name:     "Tier 1 boundary (50 kWh)",
			totalKWh: 50,
			expected: 102222,
		},
		{
			name:     "Tier 2 (60 kWh)",
			totalKWh: 60,
			expected: 123346.8,
		},
		{
			name:     "High usage (> 400 kWh)",
			totalKWh: 500,
			expected: 1463886,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			amount := calculateElectricityBill(tt.totalKWh)
			assert.InDelta(t, tt.expected, amount, 0.001)
		})
	}
}

func TestElectricityService_FetchAndSyncMonthlyUsage_RealJSONPayload(t *testing.T) {
	ctx := context.Background()

	rawJSON := []byte(`{
	  "state": "success",
	  "alert": "Sản lượng điện sử dụng",
	  "data": {
	    "soNgay": 2,
	    "tieude": "Từ 03/08/2026 đến 04/08/2026",
	    "sanluong_tungngay": [
	      {
	        "ngay": "03/08",
	        "ngayFull": "03/08/2026",
	        "sanluong_tong": "11.40",
	        "thoidiemdo": "03/08/2026"
	      },
	      {
	        "ngay": "04/08",
	        "ngayFull": "04/08/2026",
	        "sanluong_tong": "13.10",
	        "thoidiemdo": "04/08/2026"
	      }
	    ]
	  }
	}`)

	var evnResp response.DailyPowerUsageResponse
	err := json.Unmarshal(rawJSON, &evnResp)
	require.NoError(t, err)

	evnMock := &mockEVNClient{
		response: &evnResp,
	}

	repositoryMock := &mockElectricityRepository{}
	svc := NewElectricityService(repositoryMock, evnMock)

	result, err := svc.FetchAndSyncMonthlyUsage(ctx, request.DailyPowerUsageRequest{})
	require.NoError(t, err)
	require.NotNil(t, result)

	assert.InDelta(t, 24.5, result.TotalKWh, 0.001)
	require.Len(t, repositoryMock.consumptions, 2)

	assert.Equal(t, time.Date(2026, time.August, 3, 0, 0, 0, 0, time.UTC), repositoryMock.consumptions[0].MeasurementDate)
	assert.InDelta(t, 11.40, repositoryMock.consumptions[0].ConsumptionKWh, 0.001)

	assert.Equal(t, time.Date(2026, time.August, 4, 0, 0, 0, 0, time.UTC), repositoryMock.consumptions[1].MeasurementDate)
	assert.InDelta(t, 13.10, repositoryMock.consumptions[1].ConsumptionKWh, 0.001)
}
