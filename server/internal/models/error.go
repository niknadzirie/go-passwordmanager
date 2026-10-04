// Package models error response model
package models

// ErrorResponse struct
type ErrorResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
