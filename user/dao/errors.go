package dao

import "fmt"

// ErrUserNotFound is returned when a user for the provided ID was not found
type ErrUserNotFound string

func (e ErrUserNotFound) Error() string {
	return fmt.Sprintf("user not found with ID %s", string(e))
}

// ErrPictureNotFound is returned when a picture for the provided ID was not found
type ErrPictureNotFound string

func (e ErrPictureNotFound) Error() string {
	return fmt.Sprintf("picture not found with ID %s", string(e))
}
