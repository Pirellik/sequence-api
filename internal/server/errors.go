package server

import "net/http"

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (e *APIError) Error() string {
	return e.Message
}

func (e *APIError) StatusCode() int {
	return e.Code
}

func ErrInternal(msg string) error {
	return &APIError{
		Code:    http.StatusInternalServerError,
		Message: msg,
	}
}

func ErrNotFound(msg string) error {
	return &APIError{
		Code:    http.StatusNotFound,
		Message: msg,
	}
}

func ErrBadRequest(msg string) error {
	return &APIError{
		Code:    http.StatusBadRequest,
		Message: msg,
	}
}
