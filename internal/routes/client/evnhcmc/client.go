package evnhcmc

import (
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"time"
)

const (
	baseURL     = ""
	loginURL    = baseURL + ""
	dienNangURL = baseURL + ""
)

type EVNClient struct {
	httpClient *http.Client
	baseURL    *url.URL
}

func NewEVNClient() (*EVNClient, error) {
	jar, err := cookiejar.New(nil)
	if err != nil {
		return nil, err
	}

	parsedBaseURL, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}

	return &EVNClient{
		httpClient: &http.Client{
			Timeout: 20 * time.Second,
			Jar:     jar,
		},
		baseURL: parsedBaseURL,
	}, nil
}
