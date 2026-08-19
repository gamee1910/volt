package evnhcmc

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type DailyPowerUsageRequest struct {
	Token        string
	CustomerCode string
	FromDate     string
	ToDate       string
}

type DailyPowerUsageResponse struct {
	State string              `json:"state"`
	Alert string              `json:"alert"`
	Data  DailyPowerUsageData `json:"data"`
}

type DailyPowerUsageData struct {
	NumberOfDays int               `json:"soNgay"`
	Title        string            `json:"tieude"`
	DailyOutputs []DailyPowerUsage `json:"sanluong_tungngay"`
}

type DailyPowerUsage struct {
	Date                 string  `json:"ngay"`
	FullDate             string  `json:"ngayFull"`
	OffPeakIndex         float64 `json:"TD"`
	StandardIndex        float64 `json:"BT"`
	PeakIndex            float64 `json:"CD"`
	TotalIndex           float64 `json:"Tong"`
	OffPeakOutput        string  `json:"sanluong_TD"`
	StandardOutput       string  `json:"sanluong_BT"`
	PeakOutput           string  `json:"sanluong_CD"`
	TotalOutput          string  `json:"sanluong_tong"`
	MultiplicationFactor float64 `json:"hsn"`
	ActiveStandardDeliv  string  `json:"p_giao_bt"`
	ActiveOffPeakDeliv   string  `json:"p_giao_td"`
	ActivePeakDeliv      string  `json:"p_giao_cd"`
	TotalActiveDeliv     string  `json:"tong_p_giao"`
	TotalReactiveDeliv   string  `json:"tong_q_giao"`
	MeasurementTimestamp string  `json:"thoidiemdo"`
	IsBilled             int     `json:"isChotHoaDon"`
}

func (c *EVNClient) GetDailyPowerUsageData(
	ctx context.Context,
	reqData DailyPowerUsageRequest,
) (*DailyPowerUsageResponse, error) {

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	if err := writer.WriteField("input_makh", reqData.CustomerCode); err != nil {
		return nil, err
	}

	if err := writer.WriteField("input_tungay", reqData.FromDate); err != nil {
		return nil, err
	}

	if err := writer.WriteField("input_denngay", reqData.ToDate); err != nil {
		return nil, err
	}

	if err := writer.WriteField("token", reqData.Token); err != nil {
		return nil, err
	}

	if err := writer.Close(); err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		dienNangURL,
		payload,
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set(
		"Content-Type",
		writer.FormDataContentType(),
	)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf(
			"request EVNHCMC failed: %w",
			err,
		)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf(
			"read EVNHCMC response failed: %w",
			err,
		)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf(
			"EVNHCMC returned status %d: %s",
			resp.StatusCode,
			string(body),
		)
	}

	var result DailyPowerUsageResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf(
			"parse EVNHCMC response failed: %w",
			err,
		)
	}

	return &result, nil
}
