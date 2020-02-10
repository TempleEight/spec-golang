package dao

import "fmt"

// ErrMatchNotFound is returned when a user for the provided ID was not found
type ErrMatchNotFound int64

func (e1 ErrMatchNotFound) Error() string {
	return fmt.Sprintf("match not found with ID %d", e1)
}
