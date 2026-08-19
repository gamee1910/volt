package evnhcmc

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/gamee1910/volt/internal/config"
	"github.com/gamee1910/volt/pkg/logger"
)

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
