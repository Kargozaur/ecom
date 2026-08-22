package json

import (
	"encoding/json/v2"
	"net/http"
)

func Write(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.MarshalWrite(w, data)
}

func Read(r *http.Request, data any) error {
	return json.UnmarshalRead(r.Body, data, json.RejectUnknownMembers(true))
}
