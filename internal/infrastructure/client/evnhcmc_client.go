package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/gamee1910/volt/internal/application/dto"
)

type EVNClient struct {
	httpClient                *http.Client
	baseURL                   *url.URL
	loginAPI                  string
	electricityConsumptionAPI string
}

func NewEVNClient(rawBaseURL, rawLoginAPIPath, rawElectricityConsumptionAPIPath string) (*EVNClient, error) {
	if rawBaseURL == "" {
		return nil, fmt.Errorf("EVN_BASE_URL is required")
	}
	if rawLoginAPIPath == "" {
		return nil, fmt.Errorf("EVN_PATH_LOGIN is required")
	}
	if rawElectricityConsumptionAPIPath == "" {
		return nil, fmt.Errorf("EVN_PATH_DIEN_NANG_NGAY is required")
	}

	parsedBaseURL, err := url.Parse(rawBaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse base url: %w", err)
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, fmt.Errorf("create cookie jar: %w", err)
	}

	return &EVNClient{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Jar:     jar,
		},
		baseURL:                   parsedBaseURL,
		loginAPI:                  rawLoginAPIPath,
		electricityConsumptionAPI: rawElectricityConsumptionAPIPath,
	}, nil
}

func (c *EVNClient) Login(ctx context.Context, username, password string) error {
	fields := map[string]string{
		"u":        username,
		"p":        password,
		"remember": "1",
		"token":    "",
	}

	_, err := c.postMultipart(ctx, c.loginAPI, fields)
	if err != nil {
		return fmt.Errorf("login EVNHCMC: %w", err)
	}

	return nil
}

func (c *EVNClient) GetDailyPowerUsageData(
	ctx context.Context, reqData dto.DailyPowerUsageRequest,
) (*dto.DailyPowerUsageResponse, error) {
	fields := map[string]string{
		"input_makh":    reqData.CustomerCode,
		"input_tungay":  reqData.FromDate,
		"input_denngay": reqData.ToDate,
		"token":         reqData.Token,
	}

	body, err := c.postMultipart(ctx, c.electricityConsumptionAPI, fields)
	if err != nil {
		return nil, fmt.Errorf("get daily power usage: %w", err)
	}

	// raw JSON struct để unmarshal đúng field name từ EVN API
	var raw struct {
		State string `json:"state"`
		Alert string `json:"alert"`
		Data  struct {
			NumberOfDays int    `json:"soNgay"`
			Title        string `json:"tieude"`
			DailyOutputs []struct {
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
				MeasurementTimestamp string  `json:"thoidiemdo"`
				IsBilled             int     `json:"isChotHoaDon"`
			} `json:"sanluong_tungngay"`
		} `json:"data"`
	}
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("parse daily power usage response: %w", err)
	}

	result := &dto.DailyPowerUsageResponse{
		State: raw.State,
		Alert: raw.Alert,
		Data: dto.DailyPowerUsageData{
			NumberOfDays: raw.Data.NumberOfDays,
			Title:        raw.Data.Title,
		},
	}
	for _, item := range raw.Data.DailyOutputs {
		result.Data.DailyOutputs = append(result.Data.DailyOutputs, dto.DailyPowerUsage{
			Date:                 item.Date,
			FullDate:             item.FullDate,
			OffPeakIndex:         item.OffPeakIndex,
			StandardIndex:        item.StandardIndex,
			PeakIndex:            item.PeakIndex,
			TotalIndex:           item.TotalIndex,
			OffPeakOutput:        item.OffPeakOutput,
			StandardOutput:       item.StandardOutput,
			PeakOutput:           item.PeakOutput,
			TotalOutput:          item.TotalOutput,
			MultiplicationFactor: item.MultiplicationFactor,
			MeasurementTimestamp: item.MeasurementTimestamp,
			IsBilled:             item.IsBilled,
		})
	}

	return result, nil
}

func (c *EVNClient) postMultipart(
	ctx context.Context, endpointPath string, fields map[string]string,
) ([]byte, error) {
	targetURL := c.baseURL.ResolveReference(&url.URL{Path: endpointPath}).String()

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, val := range fields {
		if err := writer.WriteField(key, val); err != nil {
			return nil, fmt.Errorf("write multipart field %s: %w", key, err)
		}
	}

	if err := writer.Close(); err != nil {
		return nil, fmt.Errorf("close multipart writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, targetURL, &buf)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, string(body))
	}

	return body, nil
}
