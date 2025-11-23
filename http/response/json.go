package response

import (
	"encoding/json"
	"net/http"
)

// SuccessResponse represents a successful API response.
type SuccessResponse struct {
	Success bool        `json:"success"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

// JSON writes a JSON response with the given status code and data.
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// Success writes a successful JSON response with optional message.
func Success(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusOK, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created writes a 201 Created response with Location header.
func Created(w http.ResponseWriter, data interface{}, location string) {
	if location != "" {
		w.Header().Set("Location", location)
	}
	JSON(w, http.StatusCreated, SuccessResponse{
		Success: true,
		Data:    data,
	})
}

// NoContent writes a 204 No Content response.
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// Accepted writes a 202 Accepted response.
func Accepted(w http.ResponseWriter, data interface{}, message string) {
	JSON(w, http.StatusAccepted, SuccessResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}
