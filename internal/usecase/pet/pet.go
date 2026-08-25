// Package pet implements the Pet use case.
package pet

import (
	"context"
	"fmt"

	"github.com/DealUnloker/mockzoo/internal/entity"
	"github.com/DealUnloker/mockzoo/internal/repo"
	"github.com/DealUnloker/mockzoo/internal/usecase"
)

const (
	_defaultLimit = 20
	_maxLimit     = 100
)

// UseCase -.
type UseCase struct {
	repo repo.PetRepo
}

// New returns a Pet usecase.
func New(r repo.PetRepo) usecase.Pet {
	return &UseCase{repo: r}
}

// List -.
func (uc *UseCase) List(ctx context.Context, filter usecase.PetFilter) ([]entity.Pet, int, error) {
	if filter.Status != nil && !filter.Status.Valid() {
		return nil, 0, fmt.Errorf("%w: invalid pet status", entity.ErrValidation)
	}

	if filter.Species != nil && !filter.Species.Valid() {
		return nil, 0, fmt.Errorf("%w: invalid pet species", entity.ErrValidation)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = _defaultLimit
	}

	if limit > _maxLimit {
		limit = _maxLimit
	}

	offset := max(filter.Offset, 0)

	pets, total, err := uc.repo.List(ctx, repo.PetFilter{
		Status:  filter.Status,
		Species: filter.Species,
		Limit:   uint64(limit),
		Offset:  uint64(offset),
	})
	if err != nil {
		return nil, 0, fmt.Errorf("PetUseCase - List - uc.repo.List: %w", err)
	}

	return pets, total, nil
}

// GetByID -.
func (uc *UseCase) GetByID(ctx context.Context, id int64) (entity.Pet, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return entity.Pet{}, fmt.Errorf("PetUseCase - GetByID - uc.repo.GetByID: %w", err)
	}

	return p, nil
}

// Create -.
//
//nolint:gocritic // input is passed by value to match the usecase.Pet interface contract
func (uc *UseCase) Create(ctx context.Context, input usecase.CreatePetInput) (entity.Pet, error) {
	if input.Name == "" {
		return entity.Pet{}, fmt.Errorf("%w: name is required", entity.ErrValidation)
	}

	if !input.Species.Valid() {
		return entity.Pet{}, fmt.Errorf("%w: species is required and must be valid", entity.ErrValidation)
	}

	status := entity.PetStatusAvailable

	if input.Status != nil {
		if !input.Status.Valid() {
			return entity.Pet{}, fmt.Errorf("%w: invalid pet status", entity.ErrValidation)
		}

		status = *input.Status
	}

	tags := input.Tags
	if tags == nil {
		tags = []string{}
	}

	p := entity.Pet{
		Name:     input.Name,
		Status:   status,
		Species:  input.Species,
		Breed:    input.Breed,
		PhotoURL: input.PhotoURL,
		Tags:     tags,
	}

	if err := uc.repo.Create(ctx, &p); err != nil {
		return entity.Pet{}, fmt.Errorf("PetUseCase - Create - uc.repo.Create: %w", err)
	}

	return p, nil
}

// Update -.
func (uc *UseCase) Update(ctx context.Context, id int64, input usecase.UpdatePetInput) (entity.Pet, error) {
	p, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return entity.Pet{}, fmt.Errorf("PetUseCase - Update - uc.repo.GetByID: %w", err)
	}

	if err := applyUpdate(&p, input); err != nil {
		return entity.Pet{}, err
	}

	if err := uc.repo.Update(ctx, &p); err != nil {
		return entity.Pet{}, fmt.Errorf("PetUseCase - Update - uc.repo.Update: %w", err)
	}

	return p, nil
}

func applyUpdate(p *entity.Pet, input usecase.UpdatePetInput) error {
	if input.Name != nil {
		if *input.Name == "" {
			return fmt.Errorf("%w: name cannot be empty", entity.ErrValidation)
		}

		p.Name = *input.Name
	}

	if input.Species != nil {
		if !input.Species.Valid() {
			return fmt.Errorf("%w: invalid pet species", entity.ErrValidation)
		}

		p.Species = *input.Species
	}

	if input.Status != nil {
		if !input.Status.Valid() {
			return fmt.Errorf("%w: invalid pet status", entity.ErrValidation)
		}

		p.Status = *input.Status
	}

	if input.Breed != nil {
		p.Breed = input.Breed
	}

	if input.PhotoURL != nil {
		p.PhotoURL = input.PhotoURL
	}

	if input.Tags != nil {
		p.Tags = *input.Tags
	}

	return nil
}

// Delete -.
func (uc *UseCase) Delete(ctx context.Context, id int64) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("PetUseCase - Delete - uc.repo.Delete: %w", err)
	}

	return nil
}

// Reset -.
func (uc *UseCase) Reset(ctx context.Context) error {
	if err := uc.repo.Reset(ctx); err != nil {
		return fmt.Errorf("PetUseCase - Reset - uc.repo.Reset: %w", err)
	}

	return nil
}
