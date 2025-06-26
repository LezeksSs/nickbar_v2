package models

type Ingredient struct {
	Id          string
	Name        IngredientNomenclature
	Amount      float32
	Measure     string
	Replacement []Ingredient
	Optional    bool
	Decorative  bool
	Position    int
}

type IngredientNomenclature struct {
	Id       string
	Name     string
	Picture  string
	UserId   string
	Approved bool
}
