package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
)

type CocktailService struct {
	cocktailRepository      CocktailRepository
	ingredientRepository    CsIngredientRepository
	tagRepository           CsTagRepository
	ingredientNomRepository CsIngredientNomRepository
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=CocktailRepository --output=../../internal/mocks --with-expecter
type CocktailRepository interface {
	CreateCocktail(ctx context.Context, cocktail models.Cocktail) (models.Cocktail, error)
	IsCocktailExistAndApproved(ctx context.Context, name string) (bool, error)
	GetCocktailByID(ctx context.Context, id string) (models.Cocktail, error)
	GetCocktails(ctx context.Context, isApproved bool, usedId string) ([]models.Cocktail, error)
	GetCocktailsByIngredients(ctx context.Context, ingredientNames []string) ([]models.Cocktail, error)
	GetCocktailsByName(ctx context.Context, name string) ([]models.Cocktail, error)
	GetCocktailsByTags(ctx context.Context, tagNames []string) ([]models.Cocktail, error)
	UpdateCocktail(ctx context.Context, cocktail models.Cocktail) error
	DeleteCocktail(ctx context.Context, id string) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=CsIngredientRepository --output=../../internal/mocks --with-expecter
type CsIngredientRepository interface {
	CreateIngredient(ctx context.Context, cocktailID string, ingredient models.Ingredient) (models.Ingredient, error)
	// DeleteIngredient(ctx context.Context, ingredientID, cocktailID, userID string) error
	GetIngredientsByCocktailID(ctx context.Context, cocktailId string) ([]models.Ingredient, error)
	// InsertIngredient(ctx context.Context, ingredient models.Ingredient, cocktailID string) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=CsTagRepository --output=../../internal/mocks --with-expecter
type CsTagRepository interface {
	CreateTag(ctx context.Context, tag models.Tag) (models.Tag, error)
	IsTagExistAndApproved(ctx context.Context, name string) (bool, models.Tag, error)
	// GetTag(ctx context.Context, name string) (models.Tag, error)
	// GetTags(ctx context.Context) ([]models.Tag, error)
	// GetPersonalTags(ctx context.Context, userId string) ([]models.Tag, error)
	// GetUnapprovedTags(ctx context.Context) ([]models.Tag, error)
	// GetApprovedTags(ctx context.Context) ([]models.Tag, error)
	// UpdateTag(ctx context.Context, tag models.Tag) error
	// UpdateTagsApprovedStatus(ctx context.Context, tags []models.Tag) error
	// DeleteTag(ctx context.Context, id string) error
	GetTagsByCocktailID(ctx context.Context, cocktailId string) ([]models.Tag, error)
	// InsertCocktailTag(ctx context.Context, cocktailId, tagId string) error
	// DeleteCocktailTag(ctx context.Context, cocktailId, tagId, userId string, userRole int) error
}

//go:generate go run github.com/vektra/mockery/v2@latest --name=CsIngredientNomRepository --output=../../internal/mocks --with-expecter
type CsIngredientNomRepository interface {
	CreateIngredientNom(ctx context.Context, nomenclature models.IngredientNomenclature) (models.IngredientNomenclature, error)
	// GetIngredientNom(ctx context.Context, name string) (models.IngredientNomenclature, error)
	// GetIngredientNoms(ctx context.Context) ([]models.IngredientNomenclature, error)
	// GetUnapprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	// GetApprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	// GetPersonalIngredientNom(ctx context.Context, userId string) ([]models.IngredientNomenclature, error)
	// UpdateIngredientNom(ctx context.Context, nomenclature models.IngredientNomenclature) error
	// UpdateIngredientNomsApprovedStatus(ctx context.Context, nomenclatures []models.IngredientNomenclature) error
	// DeleteIngredientNom(ctx context.Context, id string) error
	IsIngredientNomExistAndApproved(ctx context.Context, name string) (bool, models.IngredientNomenclature, error)
}

func NewCocktailService(cocktailRepository CocktailRepository, ingredientRepository CsIngredientRepository, tagRepository CsTagRepository, ingredientNomRepository CsIngredientNomRepository) *CocktailService {
	return &CocktailService{cocktailRepository: cocktailRepository, ingredientRepository: ingredientRepository, tagRepository: tagRepository, ingredientNomRepository: ingredientNomRepository}
}

var mapCocktail = func(userId string, data requests.CocktailRequest) models.Cocktail {
	var ingredients []models.Ingredient
	for _, ingReq := range data.Ingredients {
		ing := mapIngredient(userId, ingReq)
		ingredients = append(ingredients, ing)
	}

	var tags []models.Tag
	for _, tagReq := range data.Tags {
		tag := mapTag(userId, tagReq)
		tags = append(tags, tag)
	}

	cocktail := models.Cocktail{
		Name:        data.Name,
		Picture:     data.Picture,
		Rating:      data.Rating,
		Ingredients: ingredients,
		Description: data.Description,
		Recipe:      data.Recipe,
		Tags:        tags,
		UserId:      userId,
	}

	return cocktail
}

func mapIngredient(userId string, req requests.IngredientRequest) models.Ingredient {
	var replacements []models.Ingredient
	for _, replReq := range req.Replacement {
		repl := mapIngredient(userId, replReq)
		replacements = append(replacements, repl)
	}

	nomenclature := models.IngredientNomenclature{Name: req.Name, Picture: "", UserId: userId}

	ingredient := models.Ingredient{
		Name:        nomenclature,
		Amount:      req.Amount,
		Measure:     req.Measure,
		Replacement: replacements,
		Optional:    req.Optional,
		Decorative:  req.Decorative,
		Position:    req.Position,
	}

	return ingredient
}

func mapTag(userId string, req requests.TagRequest) models.Tag {
	tag := models.Tag{
		Name:   req.Name,
		UserId: userId,
	}
	return tag
}

func (cs *CocktailService) CreateCocktail(ctx context.Context, data requests.CocktailRequest) (models.Cocktail, error) {
	var cocktail models.Cocktail
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return models.Cocktail{}, errors.New("unathorized")
	}

	exist, err := cs.cocktailRepository.IsCocktailExistAndApproved(ctx, data.Name)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("iscocktailexist error %s", err.Error())
	}
	if exist {
		return models.Cocktail{}, errors.New("cocktail has been already created")
	}

	cocktail = mapCocktail(user.Id, data)

	// Админ создает коктейль напрямую
	// Сделать реализацию модерации для созданного коктейля не у админа (возможно поле moderated для коктейля в БД)
	// Уточнить момент с ролями и какой ID-роли у админа
	for i, val := range cocktail.Ingredients {
		exist, nom, err := cs.ingredientNomRepository.IsIngredientNomExistAndApproved(ctx, val.Name.Name)
		cocktail.Ingredients[i].Name.Id = nom.Id
		if err != nil {
			return models.Cocktail{}, fmt.Errorf("isingredientnomexist error %s", err.Error())
		}
		if !exist {
			if user.Role == 1 {
				val.Name.Approved = true
				for j := 0; j < len(cocktail.Ingredients[i].Replacement); j++ {
					cocktail.Ingredients[i].Replacement[j].Name.Approved = true
				}
			}
			obj, err1 := cs.ingredientNomRepository.CreateIngredientNom(ctx, val.Name)
			cocktail.Ingredients[i].Name.Id = obj.Id
			if err1 != nil {
				fmt.Println("CreateIngredientNom error")
				return models.Cocktail{}, fmt.Errorf("iscreateingredientnom error %s", err1.Error())
			}
		}

		for j, replacement := range val.Replacement {
			existRep, nomRep, errRep := cs.ingredientNomRepository.IsIngredientNomExistAndApproved(ctx, replacement.Name.Name)
			cocktail.Ingredients[i].Replacement[j].Name.Id = nomRep.Id
			if errRep != nil {
				return models.Cocktail{}, fmt.Errorf("GetIngredientNomIfExistAndApproved for replacement error %s", errRep.Error())
			}
			if !existRep {
				if user.Role == 1 {
					replacement.Name.Approved = true
				}
				objRep, errRep1 := cs.ingredientNomRepository.CreateIngredientNom(ctx, replacement.Name)
				cocktail.Ingredients[i].Replacement[j].Name.Id = objRep.Id
				if errRep1 != nil {
					fmt.Println("CreateIngredientNom for replacement error")
					return models.Cocktail{}, fmt.Errorf("CreateIngredientNom for replacement error %s", errRep1.Error())
				}
			}
		}
	}

	// Possible refactor
	// Есть возможность дубликации имени и не совпадения id
	for i, val := range cocktail.Tags {
		exist, tag, err := cs.tagRepository.IsTagExistAndApproved(ctx, val.Name)
		cocktail.Tags[i].Id = tag.Id
		if err != nil {
			return models.Cocktail{}, fmt.Errorf("istagexist error %s", err.Error())
		}
		if !exist {
			if user.Role == 1 {
				val.Approved = true
			}
			val.UserId = user.Id
			obj, err1 := cs.tagRepository.CreateTag(ctx, val)
			cocktail.Tags[i].Id = obj.Id
			if err1 != nil {
				fmt.Println("CreateTag error")
				return models.Cocktail{}, fmt.Errorf("istagcreate error %s", err1.Error())
			}
		}
	}

	fmt.Println("")
	js, _ := json.MarshalIndent(cocktail, "", "     ")
	fmt.Println(string(js))
	fmt.Println("")

	if user.Role == 1 {
		for i := 0; i < len(cocktail.Ingredients); i++ {
			cocktail.Ingredients[i].Name.Approved = true
			for j := 0; j < len(cocktail.Ingredients[i].Replacement); j++ {
				cocktail.Ingredients[i].Replacement[j].Name.Approved = true
				fmt.Println("REPLACEMNET  ", cocktail.Ingredients[i].Replacement[j].Name.Approved)
			}
		}

		for i := 0; i < len(cocktail.Tags); i++ {
			cocktail.Tags[i].Approved = true
		}

		cocktail.Approved = true

		cocktail, err = cs.cocktailRepository.CreateCocktail(ctx, cocktail)
		if err != nil {
			return models.Cocktail{}, fmt.Errorf("createcocktail error %s", err.Error())
		}
	} else {
		cocktail, err = cs.cocktailRepository.CreateCocktail(ctx, cocktail)
		if err != nil {
			return models.Cocktail{}, fmt.Errorf("createcocktail error %s", err.Error())
		}
	}
	fmt.Println("")
	js, _ = json.MarshalIndent(cocktail, "", "     ")
	fmt.Println(string(js))
	fmt.Println("")
	return cocktail, nil

}

func (cs *CocktailService) UpdateCocktail(ctx context.Context, data requests.CocktailRequest) error {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	cocktail := mapCocktail(user.Id, data)

	if user.Role == 1 || user.Id == cocktail.UserId {
		err := cs.cocktailRepository.UpdateCocktail(ctx, cocktail)
		if err != nil {
			return err
		}
		return nil
	}
	return errors.New("restricted access")
}

func (cs *CocktailService) GetCocktailsByName(ctx context.Context, name string) ([]models.Cocktail, error) {
	cocktails, err := cs.cocktailRepository.GetCocktailsByName(ctx, name)
	if err != nil {
		return nil, err
	}

	for i := range cocktails {
		ingredients, err := cs.ingredientRepository.GetIngredientsByCocktailID(ctx, cocktails[i].Id)
		if err != nil {
			return nil, err
		}
		cocktails[i].Ingredients = ingredients

		tags, err := cs.tagRepository.GetTagsByCocktailID(ctx, cocktails[i].Id)
		if err != nil {
			return nil, err
		}
		cocktails[i].Tags = tags
	}

	return cocktails, nil
}

func (cs *CocktailService) GetCocktails(ctx context.Context, isApproved bool) ([]models.Cocktail, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return []models.Cocktail{}, errors.New("unathorized")
	}

	cocktails, err := cs.cocktailRepository.GetCocktails(ctx, isApproved, user.Id)
	if err != nil {
		return nil, err
	}

	for i := range cocktails {
		ingredients, err := cs.ingredientRepository.GetIngredientsByCocktailID(ctx, cocktails[i].Id)
		if err != nil {
			return nil, err
		}
		cocktails[i].Ingredients = ingredients

		tags, err := cs.tagRepository.GetTagsByCocktailID(ctx, cocktails[i].Id)
		if err != nil {
			return nil, err
		}
		cocktails[i].Tags = tags
	}

	return cocktails, nil
}

// ингредиенты и теги должны прилететь полностью (со всеми зависимостями)
func (cs *CocktailService) SearchCocktails(ctx context.Context, name string, ingredients []requests.IngredientRequest, tags []requests.TagRequest, pagination models.Pagination) ([]models.Cocktail, error) {
	ingredientNames := []string{}
	for _, ingredient := range ingredients {
		ingredientNames = append(ingredientNames, ingredient.Name)
	}

	tagNames := []string{}
	for _, tag := range tags {
		tagNames = append(tagNames, tag.Name)
	}

	// Получаем коктейли, содержащие заданные ингредиенты
	cocktailsByIngredients, err := cs.cocktailRepository.GetCocktailsByIngredients(ctx, ingredientNames)
	if err != nil {
		return nil, err
	}

	// Получаем коктейли по тегам из репозитория
	cocktailsByTags, err := cs.cocktailRepository.GetCocktailsByTags(ctx, tagNames)
	if err != nil {
		return nil, err
	}

	cocktailIntersection := make(map[string]int)

	for _, v := range cocktailsByIngredients {
		cocktailIntersection[v.Id] = cocktailIntersection[v.Id] + 1
	}

	for _, v := range cocktailsByTags {
		cocktailIntersection[v.Id] = cocktailIntersection[v.Id] + 1
	}

	var cocktails []models.Cocktail
	var iter []models.Cocktail = cocktailsByIngredients
	if len(cocktailsByIngredients) >= len(cocktailsByTags) {
		iter = cocktailsByTags
	}
	for i, v := range iter {
		if cocktailIntersection[v.Id] > 1 {
			cocktails = append(cocktails, iter[i])
		}
	}

	if pagination.PageNumber < 1 {
		pagination.PageNumber = 1
	}
	if pagination.PageSize < 1 {
		pagination.PageSize = 10
	}

	totalLength := len(cocktails)
	start := (pagination.PageNumber - 1) * pagination.PageSize

	if start >= totalLength {
		return []models.Cocktail{}, nil
	}

	end := start + pagination.PageSize

	if end > totalLength {
		end = totalLength
	}

	paginatedCocktails := cocktails[start:end]

	return paginatedCocktails, nil
}

func (cs *CocktailService) DeleteCocktail(ctx context.Context, id string) error {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	cocktail, err := cs.cocktailRepository.GetCocktailByID(ctx, id)
	if err != nil {
		return err
	}

	if user.Role == 1 || user.Id == cocktail.UserId {
		err := cs.cocktailRepository.DeleteCocktail(ctx, id)
		if err != nil {
			return err
		}
		return nil
	}

	return errors.New("access restricted")
}
