package entity

import "time"

// PetStatus -.
type PetStatus string // @name entity.PetStatus

const (
	PetStatusAvailable PetStatus = "available"
	PetStatusPending   PetStatus = "pending"
	PetStatusSold      PetStatus = "sold"
)

// PetSpecies -.
type PetSpecies string // @name entity.PetSpecies

const (
	PetSpeciesDog   PetSpecies = "dog"
	PetSpeciesCat   PetSpecies = "cat"
	PetSpeciesBird  PetSpecies = "bird"
	PetSpeciesFish  PetSpecies = "fish"
	PetSpeciesOther PetSpecies = "other"
)

// Pet -.
type Pet struct {
	ID        int64      `json:"id"                  example:"1"`
	Name      string     `json:"name"                example:"Barsik"`
	Status    PetStatus  `json:"status"               example:"available"`
	Species   PetSpecies `json:"species"              example:"cat"`
	Breed     *string    `json:"breed,omitempty"      example:"domestic shorthair"`
	PhotoURL  *string    `json:"photoUrl,omitempty"   example:"https://example.com/barsik.jpg"`
	Tags      []string   `json:"tags"                 example:"fluffy,demo"`
	CreatedAt time.Time  `json:"createdAt"            example:"2026-01-01T00:00:00Z"`
} // @name entity.Pet

// Valid reports whether s is a known pet status.
func (s PetStatus) Valid() bool {
	switch s {
	case PetStatusAvailable, PetStatusPending, PetStatusSold:
		return true
	default:
		return false
	}
}

// Valid reports whether s is a known pet species.
func (s PetSpecies) Valid() bool {
	switch s {
	case PetSpeciesDog, PetSpeciesCat, PetSpeciesBird, PetSpeciesFish, PetSpeciesOther:
		return true
	default:
		return false
	}
}
