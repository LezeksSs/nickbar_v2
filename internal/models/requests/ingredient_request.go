package requests

type IngredientRequest struct {
	Name        string
	Amount      float32
	Measure     string
	Replacement []IngredientRequest
	Optional    bool
	Decorative  bool
	Position    int
}
