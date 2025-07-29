package client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
)

type CostResponse struct {
	Currency string  `json:"currency"`
	Cost     float64 `json:"cost"`
}

type CostClient struct {
	BaseURL    string
	HTTPClient *http.Client
	MaxRetries int
}

func NewCostClient(baseURL string) *CostClient {
	return &CostClient{
		BaseURL:    baseURL,
		HTTPClient: &http.Client{Timeout: 3 * time.Second},
		MaxRetries: 3,
	}
}

func (c *CostClient) GetCallCost(callID string) (*CostResponse, error) {
	var lastErr error

	for attempt := 1; attempt <= c.MaxRetries; attempt++ {
		url := fmt.Sprintf("%s/calls/%s/cost", c.BaseURL, callID)
		resp, err := c.HTTPClient.Get(url)
		if err != nil {
			lastErr = err
		} else {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				body, _ := io.ReadAll(resp.Body)
				var costResp CostResponse
				if err := json.Unmarshal(body, &costResp); err != nil {
					return nil, fmt.Errorf("error parseando respuesta de costos: %w", err)
				}
				return &costResp, nil
			}

			if resp.StatusCode == http.StatusNotFound {
				return nil, errors.New("llamada no encontrada en cost API")
			}

			lastErr = fmt.Errorf("API error status %d", resp.StatusCode)
		}

		time.Sleep(time.Duration(attempt) * time.Second) // Backoff simple
	}

	return nil, fmt.Errorf("error obteniendo costo para %s: %w", callID, lastErr)
}
