package mock

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func StartMockCostAPI() {
	rand.Seed(time.Now().UnixNano())

	http.HandleFunc("/calls/", func(w http.ResponseWriter, r *http.Request) {
		callID := extractCallID(r.URL.Path)

		switch callID {
		case "123e4567-e89b-12d3-a456-426614174999":
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Algo explotó",
				"code":    "internal_server_error",
			})
			return

		case "123e4567-e89b-12d3-a456-426614174998":
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{
				"message": "Llamada no encontrada",
				"code":    "call_not_found",
			})
			return

		default:
			cost := 3.0 + rand.Float64()*10.0 // entre 3.00 y 13.00
			currencies := []string{"ARS", "USD", "EUR"}
			currency := currencies[rand.Intn(len(currencies))]

			w.WriteHeader(http.StatusOK)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"currency": currency,
				"cost":     cost,
			})
		}
	})

	go func() {
		fmt.Println("🧪 Mock cost API escuchando en http://localhost:8080")
		http.ListenAndServe(":8080", nil)
	}()
}

func extractCallID(path string) string {
	parts := strings.Split(path, "/")
	if len(parts) >= 3 {
		return parts[2]
	}
	return "unknown"
}
