package util

import (
	"encoding/json"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, code int, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	return json.NewEncoder(w).Encode(data)
}

func WriteJSONError(w http.ResponseWriter, code int, msg string) error {
	type errorResponse struct {
		Error string `json:"error"`
	}

	return WriteJSON(w, code, &errorResponse{Error: msg})

}

func ReadJSON(w http.ResponseWriter, r *http.Request, data any) error {
	maxBytes := int64(1 << 20) // 1 MB
	r.Body = http.MaxBytesReader(w, r.Body, maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	return decoder.Decode(data)
}
