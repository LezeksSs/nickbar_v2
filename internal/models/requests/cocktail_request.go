package requests

type CocktailRequest struct {
	Name        string
	Picture     string
	Rating      float32
	Ingredients []IngredientRequest
	Description string
	Recipe      string
	Tags        []TagRequest
}
