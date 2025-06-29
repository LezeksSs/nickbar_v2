package repository

import (
	"context"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type IngredientRepository struct {
	db *pgxpool.Pool
}

func NewIngredientRepository(db *pgxpool.Pool) *IngredientRepository {
	return &IngredientRepository{db: db}
}

func (r *IngredientRepository) CreateIngredient(ctx context.Context, cocktailID string, ingredient models.Ingredient) (models.Ingredient, error) {
	const ep = "repository.ingredient.CreateIngredient"

	query := `INSERT INTO ingredients (cocktail_id, nomenclature_id, amount, measure, optional, decorative, position)
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`
	err := r.db.QueryRow(ctx, query, cocktailID, ingredient.Name.Id, ingredient.Amount, ingredient.Measure, ingredient.Optional, ingredient.Decorative, ingredient.Position).Scan(&ingredient.Id)
	if err != nil {
		return models.Ingredient{}, fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	return ingredient, err
}

func (r *IngredientRepository) GetIngredientByName(ctx context.Context, name string) (models.Ingredient, error) {
	const ep = "repository.ingredient.GetIngredientByName"

	// Implement fetching ingredient by name
	return models.Ingredient{}, errors.New("not implemented")
}

func (r *IngredientRepository) UpdateIngredient(ctx context.Context, ingredient models.Ingredient) error {
	const ep = "repository.ingredient.UpdateIngredient"

	// Implement update functionality
	return errors.New("not implemented")
}

// func (r *IngredientRepository) DeleteIngredient(ctx context.Context, id string) error {
// 	query := `DELETE FROM ingredients WHERE id=$1`
// 	_, err := r.db.Exec(ctx, query, id)
// 	return err
// }

func (r *IngredientRepository) DeleteIngredient(ctx context.Context, ingredientID, cocktailID, userID string) error {
	const ep = "repository.ingredient.DeleteIngredient"

	_, err := r.db.Exec(ctx, `
		WITH user_check AS (
			SELECT 1
			FROM cocktails c
			JOIN users u ON c.user_id = u.id
			WHERE c.id = $2 AND (u.id = $3 OR u.role = 1)
		),
		deleted_replacements AS (
			DELETE FROM ingredient_replacements
			WHERE ingredient_id = $1 OR replacement_id = $1
			RETURNING ingredient_id, replacement_id
		),
		deleted_ingredients AS (
			DELETE FROM ingredients
			WHERE id = $1 OR id IN (
				SELECT replacement_id FROM deleted_replacements
			)
			AND cocktail_id = $2
			RETURNING id
		)
		DELETE FROM ingredients
		WHERE id IN (
			SELECT id FROM deleted_ingredients
		)
		AND cocktail_id = $2
		AND EXISTS (SELECT 1 FROM user_check);
	`, ingredientID, cocktailID, userID)
	if err != nil {
		return fmt.Errorf("DELETE ERROR with permission in %s: %w", ep, err)
	}
	return nil
}

func (r *IngredientRepository) IsIngredientExist(ctx context.Context, name string) (bool, error) {
	const ep = "repository.ingredient.IsIngredientExist"

	// Implement checking if ingredient exists
	return false, errors.New("not implemented")
}

func (r *IngredientRepository) GetIngredientsByCocktailID(ctx context.Context, cocktailId string) ([]models.Ingredient, error) {
	const ep = "repository.ingredient.GetIngredientsByCocktailID"

	query := `SELECT id, nomenclature_id, amount, measure, optional, decorative, position
			  FROM ingredients WHERE cocktail_id=$1 ORDER BY position`
	rows, err := r.db.Query(ctx, query, cocktailId)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var ingredients []models.Ingredient
	for rows.Next() {
		var ing models.Ingredient
		err := rows.Scan(&ing.Id, &ing.Name.Id, &ing.Amount, &ing.Measure, &ing.Optional, &ing.Decorative, &ing.Position)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}

		// Fetch IngredientNomenclature
		ing.Name, err = r.GetIngredientNomenclatureByID(ctx, ing.Name.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH INGR NOM ERROR in %s: %w", ep, err)
		}

		// Fetch Replacements
		ing.Replacement, err = r.GetIngredientReplacements(ctx, ing.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH INGR REPLACEMENT ERROR in %s: %w", ep, err)
		}

		ingredients = append(ingredients, ing)
	}

	return ingredients, nil
}

// Additional helper methods
func (r *IngredientRepository) GetIngredientNomenclatureByID(ctx context.Context, id string) (models.IngredientNomenclature, error) {
	const ep = "repository.ingredient.GetIngredientNomenclatureByID"

	// Implement fetching IngredientNomenclature by ID
	return models.IngredientNomenclature{}, errors.New("not implemented")
}

func (r *IngredientRepository) GetIngredientReplacements(ctx context.Context, ingredientId string) ([]models.Ingredient, error) {
	const ep = "repository.ingredient.GetIngredientReplacements"

	// Implement fetching ingredient replacements
	return nil, errors.New("not implemented")
}

func (r *IngredientRepository) InsertIngredient(ctx context.Context, ingredient models.Ingredient, cocktailID string) error {
	const ep = "repository.ingredient.InsertIngredient"

	query := `INSERT INTO ingredients (cocktail_id, nomenclature_id, amount, measure, optional, decorative, position) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	err := r.db.QueryRow(ctx, query, cocktailID, ingredient.Name.Id, ingredient.Amount, ingredient.Measure, ingredient.Optional, ingredient.Decorative, ingredient.Position).
		Scan(&ingredient.Id)
	if err != nil {
		return fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	// Insert replacements
	for i := 0; i < len(ingredient.Replacement); i++ {
		err = r.insertIngredientReplacement(ctx, ingredient.Id, ingredient.Replacement[i].Id)
		if err != nil {
			return fmt.Errorf("INSERT INGR REPLACEMENT ERROR in %s: %w", ep, err)
		}
	}
	return nil
}

func (r *IngredientRepository) insertIngredientReplacement(ctx context.Context, ingredientId, replacementId string) error {
	const ep = "repository.ingredient.insertIngredientReplacement"

	query := `INSERT INTO ingredient_replacements (ingredient_id, replacement_id) VALUES ($1, $2)`
	_, err := r.db.Exec(ctx, query, ingredientId, replacementId)
	if err != nil {
		return fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	return err
}
