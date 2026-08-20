package evn

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

	"github.com/gamee1910/volt/internal/config"
	"github.com/gamee1910/volt/internal/interfaces/http/handler/request"
	"github.com/gamee1910/volt/internal/interfaces/http/handler/response"
	"github.com/gamee1910/volt/pkg/logger"
)

const (
	defaultBaseURL = ""
	loginURL       = defaultBaseURL + ""
	dienNangURL    = defaultBaseURL + ""
)

type EVNClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func NewEVNClient(baseURL *url.URL) (*EVNClient, error) {
	if baseURL == nil {
		var err error
		baseURL, err = url.Parse(defaultBaseURL)
		if err != nil {
			return nil, err
		}
	}

	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	return &EVNClient{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Jar:     jar,
		},
		baseURL: baseURL,
	}, nil
}
func (c *EVNClient) GetDailyPowerUsageData(
	ctx context.Context,
	reqData request.DailyPowerUsageRequest,
) (*response.DailyPowerUsageResponse, error) {

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

	var result response.DailyPowerUsageResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf(
			"parse EVNHCMC response failed: %w",
			err,
		)
	}

	return &result, nil
}

func (c *EVNClient) Login(ctx context.Context) error {
	envConfig := config.Get().EnvConfig

	payload := &bytes.Buffer{}
	writer := multipart.NewWriter(payload)

	if err := writer.WriteField("u", envConfig.Username); err != nil {
		return err
	}

	if err := writer.WriteField("p", envConfig.Password); err != nil {
		return err
	}

	if err := writer.WriteField("remember", "1"); err != nil {
		return err
	}

	if err := writer.WriteField("token", ""); err != nil {
		return err
	}

	if err := writer.Close(); err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		loginURL,
		payload,
	)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.httpClient.Do(req)
	if err != nil {
		logger.Error("Gửi request login EVNHCMC thất bại", "error", err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	logger.Info(
		"EVNHCMC login response",
		"status",
		resp.StatusCode,
	)

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf(
			"EVNHCMC login failed: status=%d body=%s",
			resp.StatusCode,
			string(body),
		)
	}

	return nil
}
