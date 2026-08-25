// Package usecase implements application business logic. Each logic group in own file.
package usecase

import (
	"context"

	"github.com/DealUnloker/mockzoo/internal/entity"
)

type (
	// Pet -.
	Pet interface {
		List(ctx context.Context, filter PetFilter) ([]entity.Pet, int, error)
		GetByID(ctx context.Context, id int64) (entity.Pet, error)
		Create(ctx context.Context, input CreatePetInput) (entity.Pet, error)
		Update(ctx context.Context, id int64, input UpdatePetInput) (entity.Pet, error)
		Delete(ctx context.Context, id int64) error
		Reset(ctx context.Context) error
	}

	// PetFilter -.
	PetFilter struct {
		Status  *entity.PetStatus
		Species *entity.PetSpecies
		Limit   int
		Offset  int
	}

	// CreatePetInput -.
	CreatePetInput struct {
		Name     string
		Species  entity.PetSpecies
		Status   *entity.PetStatus
		Breed    *string
		PhotoURL *string
		Tags     []string
	}

	// UpdatePetInput -. All fields are optional; nil means "leave unchanged".
	UpdatePetInput struct {
		Name     *string
		Species  *entity.PetSpecies
		Status   *entity.PetStatus
		Breed    *string
		PhotoURL *string
		Tags     *[]string
	}
)
