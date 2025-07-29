package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"phonecall-cost-processor-service/internal/domain/model"
	"phonecall-cost-processor-service/internal/domain/port/client"
)

type HTTPCostClient struct {
	baseURL string
}

func NewHTTPCostClient(baseURL string) *HTTPCostClient {
	return &HTTPCostClient{baseURL: baseURL}
}

var _ client.CostClient = (*HTTPCostClient)(nil)

func (c *HTTPCostClient) GetCallCost(callID string) (model.CostResponse, error) {
	url := fmt.Sprintf("%s/calls/%s/cost", c.baseURL, callID)
	resp, err := http.Get(url)
	if err != nil {
		return model.CostResponse{}, fmt.Errorf("error llamando a API de costos: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return model.CostResponse{}, fmt.Errorf("API error status %d", resp.StatusCode)
	}

	var parsed model.CostResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return model.CostResponse{}, fmt.Errorf("error parseando respuesta de costos: %w", err)
	}

	return parsed, nil
}
