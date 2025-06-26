package models

import "time"

//Cocktail - сущность коктейля для работы с бизнес-логикой приложения
type Cocktail struct {
	Id           string
	Name         string
	Picture      string
	Rating       float32
	Ingredients  []Ingredient
	Description  string
	Recipe       string
	Tags         []Tag
	UserId       string
	Approved     bool
	CreationDate time.Time
	UpdateDate   time.Time
}
