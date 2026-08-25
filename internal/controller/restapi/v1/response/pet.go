package response

import "github.com/DealUnloker/mockzoo/internal/entity"

// PetList -.
type PetList struct {
	Items []entity.Pet `json:"items"`
	Total int          `json:"total" example:"3"`
} // @name v1.PetList

// ResetResult -.
type ResetResult struct {
	Reset bool `json:"reset" example:"true"`
	Pets  int  `json:"pets"  example:"3"`
} // @name v1.ResetResult
