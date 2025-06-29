package service

import (
	"context"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
)

type IngredientService struct {
	ingredientRepository    IngredientRepository
	ingredientNomRepository IngrtNomRepository
}

type IngredientRepository interface {
	CreateIngredient(ctx context.Context, cocktailID string, ingredient models.Ingredient) (models.Ingredient, error)
	DeleteIngredient(ctx context.Context, ingredientID, cocktailID, userID string) error
	GetIngredientsByCocktailID(ctx context.Context, cocktailId string) ([]models.Ingredient, error)
	InsertIngredient(ctx context.Context, ingredient models.Ingredient, cocktailID string) error
}

type IngrtNomRepository interface {
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

func NewIngredientService(ingredientRepository IngredientRepository, ingredientNomRepository IngrtNomRepository) *IngredientService {
	return &IngredientService{ingredientRepository: ingredientRepository, ingredientNomRepository: ingredientNomRepository}
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

func (is *IngredientService) CreateIngredient(ctx context.Context, cocktailId string, request requests.IngredientRequest) (models.Ingredient, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return models.Ingredient{}, errors.New("unathorized")
	}
	mappedIngr := mapIngredient(user.Id, request)

	exist, nom, err := is.ingredientNomRepository.IsIngredientNomExistAndApproved(ctx, mappedIngr.Name.Name)
	mappedIngr.Name.Id = nom.Id
	if err != nil {
		return models.Ingredient{}, fmt.Errorf("isingredientnomexist error %s", err.Error())
	}
	if !exist {
		if user.Role == 1 {
			mappedIngr.Name.Approved = true
			for j := 0; j < len(mappedIngr.Replacement); j++ {
				mappedIngr.Replacement[j].Name.Approved = true
			}
		}
		obj, err1 := is.ingredientNomRepository.CreateIngredientNom(ctx, mappedIngr.Name)
		mappedIngr.Name.Id = obj.Id
		if err1 != nil {
			fmt.Println("CreateIngredientNom error")
			return models.Ingredient{}, fmt.Errorf("iscreateingredientnom error %s", err1.Error())
		}
	}

	for j, replacement := range mappedIngr.Replacement {
		existRep, nomRep, errRep := is.ingredientNomRepository.IsIngredientNomExistAndApproved(ctx, replacement.Name.Name)
		mappedIngr.Replacement[j].Name.Id = nomRep.Id
		if errRep != nil {
			return models.Ingredient{}, fmt.Errorf("GetIngredientNomIfExistAndApproved for replacement error %s", errRep.Error())
		}
		if !existRep {
			if user.Role == 1 {
				replacement.Name.Approved = true
			}
			objRep, errRep1 := is.ingredientNomRepository.CreateIngredientNom(ctx, replacement.Name)
			mappedIngr.Replacement[j].Name.Id = objRep.Id
			if errRep1 != nil {
				fmt.Println("CreateIngredientNom for replacement error")
				return models.Ingredient{}, fmt.Errorf("CreateIngredientNom for replacement error %s", errRep1.Error())
			}
		}
	}
	return mappedIngr, nil
}

func (is *IngredientService) AddIngredientToCocktail(ctx context.Context, cocktailId string, request requests.IngredientRequest) (models.Ingredient, error) {
	ingredient, err := is.CreateIngredient(ctx, cocktailId, request)
	if err != nil {
		return models.Ingredient{}, fmt.Errorf("create ingredient in AddIngredientToCocktail error %s", err.Error())
	}
	err = is.ingredientRepository.InsertIngredient(ctx, ingredient, cocktailId)
	if err != nil {
		return models.Ingredient{}, fmt.Errorf("insert ingredient in AddIngredientToCocktail error %s", err.Error())
	}
	return ingredient, nil
}

func (is *IngredientService) DeleteIngredientFromCocktail(ctx context.Context, cocktailId, ingredientId string) error {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	err := is.ingredientRepository.DeleteIngredient(ctx, ingredientId, cocktailId, user.Id)
	if err != nil {
		return fmt.Errorf("failed to delete ingredients in DeleteIngredientFromCocktail, status: %w", err)
	}

	return nil
}

func (is *IngredientService) GetIngredientsByCocktailID(ctx context.Context, cocktailId string) ([]models.Ingredient, error) {

	ingredients, err := is.ingredientRepository.GetIngredientsByCocktailID(ctx, cocktailId)
	if err != nil {
		return []models.Ingredient{}, fmt.Errorf("get ingredients in InsertCocktailTag error %s", err.Error())
	}

	return ingredients, nil
}
