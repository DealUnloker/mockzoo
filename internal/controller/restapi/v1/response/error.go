package response

// Error is the standard error envelope for every failing response.
type Error struct {
	Error ErrorBody `json:"error"`
} // @name v1.Error

// ErrorBody -.
type ErrorBody struct {
	Code    string `json:"code"    example:"pet_not_found"`
	Message string `json:"message" example:"pet not found"`
} // @name v1.ErrorBody
