package dao

import (
	"errors"
	"fmt"
)

// ErrUserNotFound is returned when a user for the provided ID was not found
type ErrUserNotFound int64

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found with ID %d", e)
}

// ErrDuplicateAuth is returned when an auth already exists
var ErrDuplicateAuth = errors.New("auth already exists")
