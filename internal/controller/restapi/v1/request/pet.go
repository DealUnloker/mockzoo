package request

import "github.com/DealUnloker/mockzoo/internal/entity"

// CreatePet -.
type CreatePet struct {
	Name     string            `json:"name"               validate:"required,max=255"`
	Species  entity.PetSpecies `json:"species"             validate:"required,oneof=dog cat bird fish other"`
	Status   *entity.PetStatus `json:"status,omitempty"    validate:"omitempty,oneof=available pending sold"`
	Breed    *string           `json:"breed,omitempty"     validate:"omitempty,max=255"`
	PhotoURL *string           `json:"photoUrl,omitempty"  validate:"omitempty,max=2048"`
	Tags     []string          `json:"tags,omitempty"      validate:"omitempty,dive,max=64"`
} // @name v1.CreatePet

// UpdatePet -. All fields are optional; a missing field leaves the current value unchanged.
type UpdatePet struct {
	Name     *string            `json:"name,omitempty"      validate:"omitempty,max=255"`
	Species  *entity.PetSpecies `json:"species,omitempty"  validate:"omitempty,oneof=dog cat bird fish other"`
	Status   *entity.PetStatus  `json:"status,omitempty"    validate:"omitempty,oneof=available pending sold"`
	Breed    *string            `json:"breed,omitempty"     validate:"omitempty,max=255"`
	PhotoURL *string            `json:"photoUrl,omitempty"  validate:"omitempty,max=2048"`
	Tags     *[]string          `json:"tags,omitempty"      validate:"omitempty,dive,max=64"`
} // @name v1.UpdatePet
