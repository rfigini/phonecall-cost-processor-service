package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/sony/gobreaker"

	"phonecall-cost-processor-service/internal/domain/model"
)

// HTTPCostClient es un cliente resiliente para la API de costos
// implementa retries exponenciales y circuit breaker.
type HTTPCostClient struct {
	baseURL string
	client  *http.Client
	cb      *gobreaker.CircuitBreaker
}

// NewHTTPCostClient inicializa el cliente con backoff y circuit breaker.
func NewHTTPCostClient(baseURL string) *HTTPCostClient {
	settings := gobreaker.Settings{
		Name:        "CostAPI-CB",
		MaxRequests: 5,
		Interval:    60 * time.Second,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 3
		},
	}
	return &HTTPCostClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		cb: gobreaker.NewCircuitBreaker(settings),
	}
}

// GetCallCost obtiene el costo de una llamada con retries y circuit breaker.
func (c *HTTPCostClient) GetCallCost(callID string) (model.CostResponse, error) {
	var result model.CostResponse
	url := fmt.Sprintf("%s/calls/%s/cost", strings.TrimRight(c.baseURL, "/"), callID)

	op := func() error {
		resp, err := c.client.Get(url)
		if err != nil {
			// error de transporte, reintentar
			return err
		}
		defer resp.Body.Close()

		// Intentar parsear body de error si no es 200
		var errResp struct {
			Message string `json:"message"`
			Code    string `json:"code"`
		}
		errDecode := json.NewDecoder(resp.Body).Decode(&errResp)

		switch resp.StatusCode {
		case http.StatusOK:
			// parsear JSON de respuesta
			var body struct {
				Cost     float64 `json:"cost"`
				Currency string  `json:"currency"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				// formato inválido, no reintentar
				return backoff.Permanent(fmt.Errorf("invalid response format: %w", err))
			}
			result.Cost = body.Cost
			result.Currency = body.Currency
			return nil

		case http.StatusNotFound:
			// llamada no existe: error permanente
			if errDecode == nil {
				return backoff.Permanent(fmt.Errorf("%s (%s)", errResp.Message, errResp.Code))
			}
			return backoff.Permanent(fmt.Errorf("call %s not found", callID))

		default:
			if resp.StatusCode >= 500 {
				// errores 5xx: reintentar
				if errDecode == nil {
					return fmt.Errorf("%s (%s)", errResp.Message, errResp.Code)
				}
				return fmt.Errorf("server error: %d", resp.StatusCode)
			}
			// otros 4xx: no reintentar
			if errDecode == nil {
				return backoff.Permanent(fmt.Errorf("%s (%s)", errResp.Message, errResp.Code))
			}
			return backoff.Permanent(fmt.Errorf("unexpected status: %d", resp.StatusCode))
		}
	}

	// Circuit breaker envuelve la ejecución con backoff
	_, err := c.cb.Execute(func() (interface{}, error) {
		expBackoff := backoff.NewExponentialBackOff()
		expBackoff.InitialInterval = 500 * time.Millisecond
		expBackoff.MaxInterval = 5 * time.Second
		expBackoff.MaxElapsedTime = 30 * time.Second
		if err := backoff.Retry(op, expBackoff); err != nil {
			return nil, err
		}
		return nil, nil
	})
	if err != nil {
		return result, err
	}

	return result, nil
}
