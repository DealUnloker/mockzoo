package entity

import "errors"

var (
	// ErrPetNotFound is returned when a pet does not exist.
	ErrPetNotFound = errors.New("pet not found")
	// ErrValidation is returned when input fails domain-level validation. Wrap it with
	// fmt.Errorf("%w: <detail>", ErrValidation) to attach a specific reason.
	ErrValidation = errors.New("validation failed")
)
