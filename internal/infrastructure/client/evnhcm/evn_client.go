package evnhcm

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

	"github.com/gamee1910/volt/internal/interfaces/restapi/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/restapi/handler/response"
)

type EVNClient struct {
	httpClient       *http.Client
	baseURL          *url.URL
	pathLogin        string
	pathDienNangNgay string
}

func NewEVNClient(rawBaseURL, rawPathLogin, rawPathDienNangNgay string) (*EVNClient, error) {
	if rawBaseURL == "" {
		return nil, fmt.Errorf("EVN_BASE_URL is required")
	}
	if rawPathLogin == "" {
		return nil, fmt.Errorf("EVN_PATH_LOGIN is required")
	}
	if rawPathDienNangNgay == "" {
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
		baseURL:          parsedBaseURL,
		pathLogin:        rawPathLogin,
		pathDienNangNgay: rawPathDienNangNgay,
	}, nil
}

func (c *EVNClient) Login(ctx context.Context, username, password string) error {
	fields := map[string]string{
		"u":        username,
		"p":        password,
		"remember": "1",
		"token":    "",
	}

	_, err := c.postMultipart(ctx, c.pathLogin, fields)
	if err != nil {
		return fmt.Errorf("login EVNHCMC: %w", err)
	}

	return nil
}

func (c *EVNClient) GetDailyPowerUsageData(ctx context.Context, reqData request.DailyPowerUsageRequest) (*response.DailyPowerUsageResponse, error) {
	fields := map[string]string{
		"input_makh":    reqData.CustomerCode,
		"input_tungay":  reqData.FromDate,
		"input_denngay": reqData.ToDate,
		"token":         reqData.Token,
	}

	body, err := c.postMultipart(ctx, c.pathDienNangNgay, fields)
	if err != nil {
		return nil, fmt.Errorf("get daily power usage: %w", err)
	}

	var result response.DailyPowerUsageResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("parse daily power usage response: %w", err)
	}

	return &result, nil
}

func (c *EVNClient) postMultipart(ctx context.Context, endpointPath string, fields map[string]string) ([]byte, error) {
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
