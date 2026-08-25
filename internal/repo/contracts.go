// Package repo implements application outer layer logic. Each logic group in own file.
package repo

import (
	"context"

	"github.com/DealUnloker/mockzoo/internal/entity"
)

type (
	// PetRepo -.
	PetRepo interface {
		List(ctx context.Context, filter PetFilter) ([]entity.Pet, int, error)
		GetByID(ctx context.Context, id int64) (entity.Pet, error)
		Create(ctx context.Context, pet *entity.Pet) error
		Update(ctx context.Context, pet *entity.Pet) error
		Delete(ctx context.Context, id int64) error
		Reset(ctx context.Context) error
	}

	// PetFilter -.
	PetFilter struct {
		Status  *entity.PetStatus
		Species *entity.PetSpecies
		Limit   uint64
		Offset  uint64
	}
)
