package requests

import "nickbar_v2/internal/models/models"

type SearchRequest struct {
	Name        string
	Ingredients []IngredientRequest
	Tags        []TagRequest
	Pagination  models.Pagination
}
