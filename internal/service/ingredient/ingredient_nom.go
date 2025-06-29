package service

import (
	"context"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
)

type IngredientNomService struct {
	ingredientNomRepository IngredientNomRepository
}

type IngredientNomRepository interface {
	CreateIngredientNom(ctx context.Context, nomenclature models.IngredientNomenclature) (models.IngredientNomenclature, error)
	// GetIngredientNom(ctx context.Context, name string) (models.IngredientNomenclature, error)
	GetIngredientNoms(ctx context.Context) ([]models.IngredientNomenclature, error)
	GetUnapprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	// GetApprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	GetPersonalIngredientNom(ctx context.Context, userId string) ([]models.IngredientNomenclature, error)
	// UpdateIngredientNom(ctx context.Context, nomenclature models.IngredientNomenclature) error
	UpdateIngredientNomsApprovedStatus(ctx context.Context, nomenclatures []models.IngredientNomenclature) error
	DeleteIngredientNom(ctx context.Context, id string) error
	// IsIngredientNomExistAndApproved(ctx context.Context, name string) (bool, models.IngredientNomenclature, error)
}

func NewIngredientNomService(ingredientNomRepository IngredientNomRepository) *IngredientNomService {
	return &IngredientNomService{ingredientNomRepository: ingredientNomRepository}
}

func (is *IngredientNomService) CreateIngredientNom(ctx context.Context, nomenclature requests.IngredientNomRequest) (models.IngredientNomenclature, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return models.IngredientNomenclature{}, errors.New("unathorized")
	}
	nom, err := is.ingredientNomRepository.CreateIngredientNom(ctx, models.IngredientNomenclature{Name: nomenclature.Name, Picture: nomenclature.Picture, UserId: user.Id})
	if err != nil {
		return models.IngredientNomenclature{}, fmt.Errorf("createingredientnom in IngredientNom error %s", err.Error())
	}
	return nom, nil
}

func (is *IngredientNomService) GetIngredientNoms(ctx context.Context) ([]models.IngredientNomenclature, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return []models.IngredientNomenclature{}, errors.New("unathorized")
	}
	if user.Role == 1 {
		nomenclatures, err := is.ingredientNomRepository.GetIngredientNoms(ctx)
		if err != nil {
			return []models.IngredientNomenclature{}, fmt.Errorf("getnomenclatures in IngredientNomService error %s", err.Error())
		}
		return nomenclatures, nil
	}
	return []models.IngredientNomenclature{}, errors.New("restricted access")
}

func (is *IngredientNomService) GetUnapprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unathorized")
	}
	if user.Role == 1 {
		nomenclatures, err := is.ingredientNomRepository.GetUnapprovedIngredientNom(ctx)
		if err != nil {
			return []models.IngredientNomenclature{}, fmt.Errorf("getunapprovednomenclatures in IngredientNomService error %s", err.Error())
		}
		return nomenclatures, nil
	}
	return []models.IngredientNomenclature{}, errors.New("restricted access")
}

func (is *IngredientNomService) GetApprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error) {
	nomenclatures, err := is.ingredientNomRepository.GetUnapprovedIngredientNom(ctx)
	if err != nil {
		return []models.IngredientNomenclature{}, fmt.Errorf("getapprovednomenclatures in IngredientNomService error %s", err.Error())
	}
	return nomenclatures, nil
}

func (is *IngredientNomService) GetPersonalIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return []models.IngredientNomenclature{}, errors.New("unathorized")
	}
	nomenclatures, err := is.ingredientNomRepository.GetPersonalIngredientNom(ctx, user.Id)
	if err != nil {
		return []models.IngredientNomenclature{}, fmt.Errorf("getpersonalnoms in IngredientNomService error %s", err.Error())
	}
	return nomenclatures, nil
}

func (is *IngredientNomService) UpdateIngredientNomApprovedStatus(ctx context.Context, ingredientNomRequest []requests.IngredientNomUpdateRequest) error {
	var nomenclatures []models.IngredientNomenclature
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	if user.Role == 1 {
		for _, req := range ingredientNomRequest {
			nomenclature := models.IngredientNomenclature{
				Id:       req.Id,
				Picture:  req.Picture,
				Approved: req.Approved, // Меняем только поле Approved
			}
			nomenclatures = append(nomenclatures, nomenclature)
		}

		err := is.ingredientNomRepository.UpdateIngredientNomsApprovedStatus(ctx, nomenclatures)
		if err != nil {
			return fmt.Errorf("failed to update nomenclatures in IngredientNomService approved status: %w", err)
		}
		return nil
	}

	return errors.New("restricted access")
}

func (is *IngredientNomService) DeleteIngredientNom(ctx context.Context, id string) error {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	if user.Role == 1 {
		err := is.ingredientNomRepository.DeleteIngredientNom(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete nomenclature in IngredientNomService, status: %w", err)
		}
		return nil
	}

	return errors.New("restricted access")
}
