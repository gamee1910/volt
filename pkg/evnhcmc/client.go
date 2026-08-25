package evnhcmc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"

	"github.com/gamee1910/volt/config"
	pkgjson "github.com/gamee1910/volt/pkg/json"
	"github.com/gamee1910/volt/pkg/logger"
)

type evnClient struct {
	httpClient                *http.Client
	baseURL                   *url.URL
	cfg                       *config.Configuration
	log                       *logger.Logger
	loginAPI                  string
	electricityConsumptionAPI string
}

func NewEVNClient(
	cfg *config.Configuration,
	log *logger.Logger,
) (EVNClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("application config is required")
	}

	var (
		rawBaseURL                       = cfg.ApplicationConfig.EVNHCMCConfig.BaseURL
		rawLoginAPIPath                  = cfg.ApplicationConfig.EVNHCMCConfig.LoginAPI
		rawElectricityConsumptionAPIPath = cfg.ApplicationConfig.EVNHCMCConfig.ElectricityConsumptionAPI
	)

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

	return &evnClient{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Jar:     jar,
		},
		baseURL:                   parsedBaseURL,
		loginAPI:                  rawLoginAPIPath,
		electricityConsumptionAPI: rawElectricityConsumptionAPIPath,
		cfg:                       cfg,
		log:                       log,
	}, nil
}

func (c *evnClient) GetDailyPowerUsageData(
	ctx context.Context, reqData DailyPowerUsageRequest,
) (*DailyPowerUsageResponse, error) {

	if err := c.ensureLogin(ctx); err != nil {
		return nil, err
	}

	return c.fetchDailyPowerUsage(ctx, reqData)
}

func (c *evnClient) login(ctx context.Context) error {
	fields := map[string]string{
		"u":        c.cfg.ApplicationConfig.EVNHCMCConfig.Username,
		"p":        c.cfg.ApplicationConfig.EVNHCMCConfig.Password,
		"remember": "1",
		"token":    "",
	}

	if _, err := c.postMultipart(ctx, c.loginAPI, fields); err != nil {
		return fmt.Errorf("login EVNHCMC: %w", err)
	}

	if !c.hasSessionCookie() {
		return fmt.Errorf("login EVNHCMC succeeded but session cookie was not created")
	}

	return nil
}

func (c *evnClient) ensureLogin(ctx context.Context) error {
	if c.hasSessionCookie() {
		return nil
	}

	if err := c.login(ctx); err != nil {
		return fmt.Errorf("login EVNHCMC: %w", err)
	}

	return nil
}

func (c *evnClient) hasSessionCookie() bool {
	cookies := c.httpClient.Jar.Cookies(c.baseURL)

	var (
		hasEVNSession bool
		hasBIGIP      bool
		hasTS         bool
	)

	for _, cookie := range cookies {
		if cookie.Value == "" {
			continue
		}

		switch cookie.Name {
		case "evn_session":
			hasEVNSession = true

		case "BIGipServerPool_CSKH_WEB_192.168.36.113_443":
			hasBIGIP = true

		case "TS018cfa5d":
			hasTS = true
		}
	}

	return hasEVNSession && hasBIGIP && hasTS
}

func (c *evnClient) fetchDailyPowerUsage(
	ctx context.Context, reqData DailyPowerUsageRequest,
) (*DailyPowerUsageResponse, error) {
	fields := map[string]string{
		"input_makh":    c.cfg.ApplicationConfig.EVNHCMCConfig.CustomerCode,
		"input_tungay":  reqData.FromDate,
		"input_denngay": reqData.ToDate,
		"token":         "",
	}

	body, err := c.postMultipart(ctx, c.electricityConsumptionAPI, fields)
	if err != nil {
		return nil, fmt.Errorf("get daily power usage: %w", err)
	}

	var response DailyPowerUsageResponse
	if err := pkgjson.Decode(body, &response); err != nil {
		return nil, fmt.Errorf("parse daily power usage response: %w", err)
	}

	return &response, nil
}

func (c *evnClient) postMultipart(
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
