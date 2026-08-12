package health

import (
	"net/http"
	"pkg/json"
)

func Health(mux *http.ServeMux) {
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		json.Write(w, http.StatusOK, map[string]string{"response": "Hello"})
	})
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		json.Write(w, http.StatusOK, map[string]string{"status": "Healthy"})
	})
}
