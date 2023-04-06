package telegram

import "fmt"

type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("Internal Error, message: %s", e.Message)
}

type APIError struct {
	ErrorCode   int    `json:"error_code"`
	Description string `json:"description"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("APIError Error, message: %s", e.Description)
}
