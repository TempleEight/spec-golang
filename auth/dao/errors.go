package dao

import "errors"

// ErrAuthNotFound is returned when an for the provided email was not found
var ErrAuthNotFound = errors.New("auth not found")
