package server

import (
	"encoding/json"
	"log"
	"net/http"
)

const (
	ErrCodeInvalidInput  = "INVALID_INPUT"
	ErrCodeInternalError = "INTERNAL_ERROR"
)

type ErrorResponse struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func invalidInput(respWriter http.ResponseWriter, message string) {
	writeError(respWriter, http.StatusBadRequest, ErrCodeInvalidInput, message)
}

func internalError(respWriter http.ResponseWriter, message string) {
	writeError(respWriter, http.StatusInternalServerError, ErrCodeInternalError, message)
}

func writeError(respWriter http.ResponseWriter, status int, code, message string) {
	respWriter.Header().Set("Content-Type", "application/json")
	respWriter.WriteHeader(status)
	writeJSONResponse(respWriter, ErrorResponse{
		Code:    code,
		Message: message,
	})
}

func writeJSONResponse(respWriter http.ResponseWriter, v any) {
	respWriter.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(respWriter).Encode(v); err != nil {
		log.Printf("failed to write response: %v", err)
	}
}
