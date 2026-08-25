package v1

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/DealUnloker/mockzoo/internal/controller/restapi/v1/request"
	"github.com/DealUnloker/mockzoo/internal/controller/restapi/v1/response"
	"github.com/DealUnloker/mockzoo/internal/entity"
	"github.com/DealUnloker/mockzoo/internal/usecase"
	"github.com/gofiber/fiber/v2"
)

const (
	_codePetNotFound      = "pet_not_found"
	_codeValidationFailed = "validation_failed"
	_codeInternal         = "internal"

	// _seedPetCount mirrors len(seedPets()) in internal/repo/persistent/pet/seed.go:
	// the sandbox always holds exactly this many pets right after a reset.
	_seedPetCount = 3
)

// handleError maps a usecase/entity error to a transport status and error envelope.
func handleError(ctx *fiber.Ctx, err error) error {
	switch {
	case errors.Is(err, entity.ErrPetNotFound):
		return errorResponse(ctx, http.StatusNotFound, _codePetNotFound, entity.ErrPetNotFound.Error())
	case errors.Is(err, entity.ErrValidation):
		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, err.Error())
	default:
		return errorResponse(ctx, http.StatusInternalServerError, _codeInternal, "internal server error")
	}
}

var errInvalidPetID = errors.New("petId must be an integer")

func parsePetID(ctx *fiber.Ctx) (int64, error) {
	id, err := strconv.ParseInt(ctx.Params("petId"), 10, 64)
	if err != nil {
		return 0, errInvalidPetID
	}

	return id, nil
}

// @Summary     List pets
// @Description List pets with optional filtering by status and species, and pagination
// @ID          listPets
// @Tags        pets
// @Produce     json
// @Param       status  query    string false "Filter by status"  Enums(available, pending, sold)
// @Param       species query    string false "Filter by species" Enums(dog, cat, bird, fish, other)
// @Param       limit   query    int    false "Max results (default 20, max 100)"
// @Param       offset  query    int    false "Results to skip"
// @Success     200     {object} response.PetList
// @Failure     400     {object} response.Error
// @Router      /pets [get]
func (r *V1) listPets(ctx *fiber.Ctx) error {
	var filter usecase.PetFilter

	if s := ctx.Query("status"); s != "" {
		status := entity.PetStatus(s)
		filter.Status = &status
	}

	if s := ctx.Query("species"); s != "" {
		species := entity.PetSpecies(s)
		filter.Species = &species
	}

	if s := ctx.Query("limit"); s != "" {
		limit, err := strconv.Atoi(s)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, "limit must be an integer")
		}

		filter.Limit = limit
	}

	if s := ctx.Query("offset"); s != "" {
		offset, err := strconv.Atoi(s)
		if err != nil {
			return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, "offset must be an integer")
		}

		filter.Offset = offset
	}

	pets, total, err := r.p.List(ctx.UserContext(), filter)
	if err != nil {
		r.l.Error(err, "restapi - v1 - listPets")

		return handleError(ctx, err)
	}

	if pets == nil {
		pets = []entity.Pet{}
	}

	return ctx.Status(http.StatusOK).JSON(response.PetList{Items: pets, Total: total})
}

// @Summary     Get a pet by id
// @ID          getPetById
// @Tags        pets
// @Produce     json
// @Param       petId path     int true "Pet ID"
// @Success     200   {object} entity.Pet
// @Failure     400   {object} response.Error "Invalid pet id"
// @Failure     404   {object} response.Error
// @Router      /pets/{petId} [get]
func (r *V1) getPetByID(ctx *fiber.Ctx) error {
	id, err := parsePetID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, err.Error())
	}

	p, err := r.p.GetByID(ctx.UserContext(), id)
	if err != nil {
		r.l.Error(err, "restapi - v1 - getPetByID")

		return handleError(ctx, err)
	}

	return ctx.Status(http.StatusOK).JSON(p)
}

// @Summary     Create a pet
// @ID          createPet
// @Tags        pets
// @Accept      json
// @Produce     json
// @Param       request body     request.CreatePet true "Pet data"
// @Success     201     {object} entity.Pet
// @Failure     400     {object} response.Error
// @Router      /pets [post]
func (r *V1) createPet(ctx *fiber.Ctx) error {
	var body request.CreatePet

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - createPet")

		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - createPet")

		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, "invalid request body")
	}

	p, err := r.p.Create(ctx.UserContext(), usecase.CreatePetInput{
		Name:     body.Name,
		Species:  body.Species,
		Status:   body.Status,
		Breed:    body.Breed,
		PhotoURL: body.PhotoURL,
		Tags:     body.Tags,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - createPet")

		return handleError(ctx, err)
	}

	return ctx.Status(http.StatusCreated).JSON(p)
}

// @Summary     Update a pet
// @Description Partially update a pet; all body fields are optional
// @ID          updatePet
// @Tags        pets
// @Accept      json
// @Produce     json
// @Param       petId   path     int               true "Pet ID"
// @Param       request body     request.UpdatePet true "Fields to update"
// @Success     200     {object} entity.Pet
// @Failure     400     {object} response.Error
// @Failure     404     {object} response.Error
// @Router      /pets/{petId} [patch]
func (r *V1) updatePet(ctx *fiber.Ctx) error {
	id, err := parsePetID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, err.Error())
	}

	var body request.UpdatePet

	if err := ctx.BodyParser(&body); err != nil {
		r.l.Error(err, "restapi - v1 - updatePet")

		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, "invalid request body")
	}

	if err := r.v.Struct(body); err != nil {
		r.l.Error(err, "restapi - v1 - updatePet")

		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, "invalid request body")
	}

	p, err := r.p.Update(ctx.UserContext(), id, usecase.UpdatePetInput{
		Name:     body.Name,
		Species:  body.Species,
		Status:   body.Status,
		Breed:    body.Breed,
		PhotoURL: body.PhotoURL,
		Tags:     body.Tags,
	})
	if err != nil {
		r.l.Error(err, "restapi - v1 - updatePet")

		return handleError(ctx, err)
	}

	return ctx.Status(http.StatusOK).JSON(p)
}

// @Summary     Delete a pet
// @ID          deletePet
// @Tags        pets
// @Param       petId path int true "Pet ID"
// @Success     204   "No Content"
// @Failure     400   {object} response.Error
// @Failure     404   {object} response.Error
// @Router      /pets/{petId} [delete]
func (r *V1) deletePet(ctx *fiber.Ctx) error {
	id, err := parsePetID(ctx)
	if err != nil {
		return errorResponse(ctx, http.StatusBadRequest, _codeValidationFailed, err.Error())
	}

	if err := r.p.Delete(ctx.UserContext(), id); err != nil {
		r.l.Error(err, "restapi - v1 - deletePet")

		return handleError(ctx, err)
	}

	return ctx.SendStatus(http.StatusNoContent)
}

// @Summary     Reset the sandbox
// @Description Restores the three canonical seed pets and discards every other pet
// @ID          resetSandbox
// @Tags        admin
// @Produce     json
// @Success     200 {object} response.ResetResult
// @Router      /admin/reset [post]
func (r *V1) resetSandbox(ctx *fiber.Ctx) error {
	if err := r.p.Reset(ctx.UserContext()); err != nil {
		r.l.Error(err, "restapi - v1 - resetSandbox")

		return handleError(ctx, err)
	}

	return ctx.Status(http.StatusOK).JSON(response.ResetResult{Reset: true, Pets: _seedPetCount})
}
