package v1

import (
	"github.com/DealUnloker/mockzoo/internal/usecase"
	"github.com/DealUnloker/mockzoo/pkg/logger"
	"github.com/go-playground/validator/v10"
)

// V1 -.
type V1 struct {
	p usecase.Pet
	l logger.Interface
	v *validator.Validate
}
