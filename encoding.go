package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func encodeError(w http.ResponseWriter, r *http.Request, status int, v string) error {
	return encode(w, r, status, struct {
		Ok    bool   `json:"ok"`
		Error string `json:"error"`
	}{
		Ok:    false,
		Error: v,
	})
}

func encodeOk[T any](w http.ResponseWriter, r *http.Request, status int, v T) error {
	return encode(w, r, status, struct {
		Ok   bool `json:"ok"`
		Data T    `json:"data"`
	}{
		Ok:   true,
		Data: v,
	})
}

func encode(w http.ResponseWriter, r *http.Request, status int, v any) error {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	err := json.NewEncoder(w).Encode(v)
	if err != nil {
		return fmt.Errorf("encode json: %w", err)
	}

	return nil
}

func decode[T any](r *http.Request) (T, error) {
	var v T

	err := json.NewDecoder(r.Body).Decode(&v)
	if err != nil {
		return v, fmt.Errorf("decode json: %w", err)
	}

	return v, nil
}
