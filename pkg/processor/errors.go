package processor

import "fmt"

type InternalError struct {
	Message string
}

func (e *InternalError) Error() string {
	return fmt.Sprintf("Internal Error, message: %s", e.Message)
}
