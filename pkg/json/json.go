package json

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Encode marshals v thành JSON. Trả về lỗi nếu marshal thất bại.
func Encode(v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json encode: %w", err)
	}
	return b, nil
}

// Decode đọc JSON từ b vào v. Trả về lỗi nếu unmarshal thất bại.
func Decode(b []byte, v any) error {
	if err := json.Unmarshal(b, v); err != nil {
		return fmt.Errorf("json decode: %w", err)
	}
	return nil
}

// WriteJSON ghi JSON response với status code cho trước.
func WriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		http.Error(w, "failed to encode response", http.StatusInternalServerError)
	}
}

// WriteOK ghi JSON response với status 200 OK.
func WriteOK(w http.ResponseWriter, v any) {
	WriteJSON(w, http.StatusOK, v)
}

// WriteError ghi JSON error response với status code và message cho trước.
func WriteError(w http.ResponseWriter, status int, message string) {
	WriteJSON(w, status, map[string]string{"error": message})
}

// DecodeBody đọc JSON từ HTTP request body vào v.
func DecodeBody(r *http.Request, v any) error {
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		return fmt.Errorf("decode request body: %w", err)
	}
	return nil
}
