package pet

import "github.com/DealUnloker/mockzoo/internal/entity"

// _seedIDStart is the sequence value Reset restores after reseeding, so the next
// user-created pet gets id 1001 and never collides with a seed row. Mirrors the
// setval() call in the seed_pets migration.
const _seedIDStart = 1000

// Stable seed pet ids, mirrored in the seed_pets migration.
const (
	_barsikID = 1
	_rexID    = 2
	_keshaID  = 3
)

// seedPets returns the three canonical sandbox pets. This is the single source of
// truth for Reset; the seed_pets migration inserts the same rows on first boot.
func seedPets() []entity.Pet {
	return []entity.Pet{
		{
			ID:      _barsikID,
			Name:    "Barsik",
			Status:  entity.PetStatusAvailable,
			Species: entity.PetSpeciesCat,
			Breed:   new("domestic shorthair"),
			Tags:    []string{"fluffy", "demo"},
		},
		{
			ID:      _rexID,
			Name:    "Rex",
			Status:  entity.PetStatusPending,
			Species: entity.PetSpeciesDog,
			Breed:   new("german shepherd"),
			Tags:    []string{"good-boy", "demo"},
		},
		{
			ID:      _keshaID,
			Name:    "Kesha",
			Status:  entity.PetStatusSold,
			Species: entity.PetSpeciesBird,
			Breed:   new("budgerigar"),
			Tags:    []string{"talks", "demo"},
		},
	}
}
