package client

import (
	"net/http"
	"net/http/httptest"
	
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGetCallCost_RetriesOnFailure(t *testing.T) {
	var attempt int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Fail the first 2 requests
		if atomic.AddInt32(&attempt, 1) < 3 {
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// Then return success
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"currency":"ARS","cost":5.75}`))
	}))
	defer ts.Close()

	c := NewHttpCostClient(ts.URL)

	start := time.Now()
	resp, err := c.GetCallCost("dummy-call-id")
	duration := time.Since(start)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "ARS", resp.Currency)
	assert.Equal(t, 5.75, resp.Cost)
	assert.GreaterOrEqual(t, int(duration.Seconds()), 1) // should wait at least 1s from backoff
	assert.Equal(t, int32(3), attempt)
}

func TestGetCallCost_FailsAfterMaxRetries(t *testing.T) {
	var attempt int32

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&attempt, 1)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}))
	defer ts.Close()

	c := NewHttpCostClient(ts.URL)

	resp, err := c.GetCallCost("failing-call-id")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Equal(t, int32(3), attempt)
}
