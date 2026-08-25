package usecase_test

import (
	"context"
	"errors"
	"testing"

	"github.com/DealUnloker/mockzoo/internal/entity"
	"github.com/DealUnloker/mockzoo/internal/repo"
	"github.com/DealUnloker/mockzoo/internal/usecase"
	"github.com/DealUnloker/mockzoo/internal/usecase/pet"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var errRepoGeneric = errors.New("repository error")

func newPetUseCase(t *testing.T) (usecase.Pet, *MockPetRepo) {
	t.Helper()

	ctrl := gomock.NewController(t)

	mockRepo := NewMockPetRepo(ctrl)
	useCase := pet.New(mockRepo)

	return useCase, mockRepo
}

func TestPetList(t *testing.T) {
	t.Parallel()

	pet1 := entity.Pet{ID: 1, Name: "Barsik", Status: entity.PetStatusAvailable, Species: entity.PetSpeciesCat}
	pet2 := entity.Pet{ID: 2, Name: "Rex", Status: entity.PetStatusPending, Species: entity.PetSpeciesDog}

	t.Run("list success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().
			List(gomock.Any(), repo.PetFilter{Limit: uint64(20), Offset: uint64(0)}).
			Return([]entity.Pet{pet1, pet2}, 2, nil)

		pets, total, err := uc.List(context.Background(), usecase.PetFilter{})

		require.NoError(t, err)
		assert.Equal(t, 2, total)
		assert.Len(t, pets, 2)
	})

	t.Run("list defaults and clamps limit", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().
			List(gomock.Any(), repo.PetFilter{Limit: uint64(100), Offset: uint64(0)}).
			Return([]entity.Pet{pet1}, 1, nil)

		pets, total, err := uc.List(context.Background(), usecase.PetFilter{Limit: 1000, Offset: -5})

		require.NoError(t, err)
		assert.Equal(t, 1, total)
		assert.Len(t, pets, 1)
	})

	t.Run("list repo error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().List(gomock.Any(), gomock.Any()).Return(nil, 0, errRepoGeneric)

		_, _, err := uc.List(context.Background(), usecase.PetFilter{})

		require.ErrorIs(t, err, errRepoGeneric)
	})
}

func TestPetListValidation(t *testing.T) {
	t.Parallel()

	t.Run("list invalid status", func(t *testing.T) {
		t.Parallel()

		uc, _ := newPetUseCase(t)

		invalid := entity.PetStatus("extinct")

		_, _, err := uc.List(context.Background(), usecase.PetFilter{Status: &invalid})

		require.ErrorIs(t, err, entity.ErrValidation)
	})

	t.Run("list invalid species", func(t *testing.T) {
		t.Parallel()

		uc, _ := newPetUseCase(t)

		invalid := entity.PetSpecies("dragon")

		_, _, err := uc.List(context.Background(), usecase.PetFilter{Species: &invalid})

		require.ErrorIs(t, err, entity.ErrValidation)
	})
}

func TestPetGetByID(t *testing.T) {
	t.Parallel()

	expected := entity.Pet{ID: 1, Name: "Barsik", Status: entity.PetStatusAvailable, Species: entity.PetSpeciesCat}

	t.Run("get success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(expected, nil)

		p, err := uc.GetByID(context.Background(), 1)

		require.NoError(t, err)
		assert.Equal(t, expected, p)
	})

	t.Run("get not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().GetByID(gomock.Any(), int64(99999)).Return(entity.Pet{}, entity.ErrPetNotFound)

		_, err := uc.GetByID(context.Background(), 99999)

		require.ErrorIs(t, err, entity.ErrPetNotFound)
	})
}

func TestPetCreate(t *testing.T) {
	t.Parallel()

	t.Run("create success with default status", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)

		p, err := uc.Create(context.Background(), usecase.CreatePetInput{
			Name:    "Barsik",
			Species: entity.PetSpeciesCat,
		})

		require.NoError(t, err)
		assert.Equal(t, "Barsik", p.Name)
		assert.Equal(t, entity.PetStatusAvailable, p.Status)
		assert.Empty(t, p.Tags)
	})

	t.Run("create missing name", func(t *testing.T) {
		t.Parallel()

		uc, _ := newPetUseCase(t)

		_, err := uc.Create(context.Background(), usecase.CreatePetInput{Species: entity.PetSpeciesCat})

		require.ErrorIs(t, err, entity.ErrValidation)
	})

	t.Run("create invalid species", func(t *testing.T) {
		t.Parallel()

		uc, _ := newPetUseCase(t)

		_, err := uc.Create(context.Background(), usecase.CreatePetInput{Name: "Barsik", Species: entity.PetSpecies("dragon")})

		require.ErrorIs(t, err, entity.ErrValidation)
	})

	t.Run("create invalid status", func(t *testing.T) {
		t.Parallel()

		uc, _ := newPetUseCase(t)

		invalid := entity.PetStatus("extinct")

		_, err := uc.Create(context.Background(), usecase.CreatePetInput{
			Name:    "Barsik",
			Species: entity.PetSpeciesCat,
			Status:  &invalid,
		})

		require.ErrorIs(t, err, entity.ErrValidation)
	})

	t.Run("create repo error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().Create(gomock.Any(), gomock.Any()).Return(errRepoGeneric)

		_, err := uc.Create(context.Background(), usecase.CreatePetInput{Name: "Barsik", Species: entity.PetSpeciesCat})

		require.ErrorIs(t, err, errRepoGeneric)
	})
}

func TestPetUpdate(t *testing.T) {
	t.Parallel()

	existing := entity.Pet{ID: 1, Name: "Barsik", Status: entity.PetStatusAvailable, Species: entity.PetSpeciesCat}

	t.Run("update success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(existing, nil)
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(nil)

		newName := "Barsik the Great"

		p, err := uc.Update(context.Background(), 1, usecase.UpdatePetInput{Name: &newName})

		require.NoError(t, err)
		assert.Equal(t, newName, p.Name)
	})

	t.Run("update not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().GetByID(gomock.Any(), int64(99999)).Return(entity.Pet{}, entity.ErrPetNotFound)

		_, err := uc.Update(context.Background(), 99999, usecase.UpdatePetInput{})

		require.ErrorIs(t, err, entity.ErrPetNotFound)
	})

	t.Run("update invalid status", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(existing, nil)

		invalid := entity.PetStatus("extinct")

		_, err := uc.Update(context.Background(), 1, usecase.UpdatePetInput{Status: &invalid})

		require.ErrorIs(t, err, entity.ErrValidation)
	})

	t.Run("update repo error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().GetByID(gomock.Any(), int64(1)).Return(existing, nil)
		mockRepo.EXPECT().Update(gomock.Any(), gomock.Any()).Return(errRepoGeneric)

		newName := "New Name"

		_, err := uc.Update(context.Background(), 1, usecase.UpdatePetInput{Name: &newName})

		require.ErrorIs(t, err, errRepoGeneric)
	})
}

func TestPetDelete(t *testing.T) {
	t.Parallel()

	t.Run("delete success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().Delete(gomock.Any(), int64(1)).Return(nil)

		err := uc.Delete(context.Background(), 1)

		require.NoError(t, err)
	})

	t.Run("delete not found", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().Delete(gomock.Any(), int64(99999)).Return(entity.ErrPetNotFound)

		err := uc.Delete(context.Background(), 99999)

		require.ErrorIs(t, err, entity.ErrPetNotFound)
	})
}

func TestPetReset(t *testing.T) {
	t.Parallel()

	t.Run("reset success", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().Reset(gomock.Any()).Return(nil)

		err := uc.Reset(context.Background())

		require.NoError(t, err)
	})

	t.Run("reset repo error", func(t *testing.T) {
		t.Parallel()

		uc, mockRepo := newPetUseCase(t)
		mockRepo.EXPECT().Reset(gomock.Any()).Return(errRepoGeneric)

		err := uc.Reset(context.Background())

		require.ErrorIs(t, err, errRepoGeneric)
	})
}
