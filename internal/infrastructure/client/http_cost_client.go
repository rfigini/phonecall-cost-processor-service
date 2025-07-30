package client

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"phonecall-cost-processor-service/internal/domain/model"
	"time"
)

type CostClient interface {
	GetCallCost(callID string) (*model.CostResponse, error)
}

type HttpCostClient struct {
	baseURL string
}

func NewHttpCostClient(baseURL string) *HttpCostClient {
	return &HttpCostClient{baseURL: baseURL}
}

func (c *HttpCostClient) GetCallCost(callID string) (*model.CostResponse, error) {
	var lastErr error
	maxRetries := 3
	backoff := time.Second

	for i := 0; i < maxRetries; i++ {
		url := fmt.Sprintf("%s/calls/%s/cost", c.baseURL, callID)
		resp, err := http.Get(url)
		if err != nil {
			lastErr = err
			log.Printf("⚠️ Error en llamada a cost API (intento %d): %v", i+1, err)
		} else {
			defer resp.Body.Close()

			if resp.StatusCode == http.StatusOK {
				var costResp model.CostResponse
				if err := json.NewDecoder(resp.Body).Decode(&costResp); err != nil {
					return nil, fmt.Errorf("error parseando respuesta de costos: %w", err)
				}
				return &costResp, nil
			}

			lastErr = fmt.Errorf("status code %d", resp.StatusCode)
			log.Printf("⚠️ Fallo en cost API (intento %d): %s", i+1, lastErr)
		}

		time.Sleep(backoff)
		backoff *= 2 // exponential
	}

	return nil, fmt.Errorf("cost API falló luego de %d intentos: %w", maxRetries, lastErr)
}
