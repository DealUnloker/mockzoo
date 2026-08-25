package entity_test

import (
	"testing"

	"github.com/DealUnloker/mockzoo/internal/entity"
	"github.com/stretchr/testify/assert"
)

func TestPetStatus_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     entity.PetStatus
		valid bool
	}{
		{"available", entity.PetStatusAvailable, true},
		{"pending", entity.PetStatusPending, true},
		{"sold", entity.PetStatusSold, true},
		{"unknown", entity.PetStatus("extinct"), false},
		{"empty", entity.PetStatus(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, tt.s.Valid())
		})
	}
}

func TestPetSpecies_Valid(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		s     entity.PetSpecies
		valid bool
	}{
		{"dog", entity.PetSpeciesDog, true},
		{"cat", entity.PetSpeciesCat, true},
		{"bird", entity.PetSpeciesBird, true},
		{"fish", entity.PetSpeciesFish, true},
		{"other", entity.PetSpeciesOther, true},
		{"unknown", entity.PetSpecies("dragon"), false},
		{"empty", entity.PetSpecies(""), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.valid, tt.s.Valid())
		})
	}
}
