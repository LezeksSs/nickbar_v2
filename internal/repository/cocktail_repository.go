package repository

import (
	"context"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"

	pgx "github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CocktailRepository struct {
	db *pgxpool.Pool
}

func NewCocktailRepository(db *pgxpool.Pool) *CocktailRepository {
	return &CocktailRepository{db: db}
}

func (r *CocktailRepository) CreateCocktail(ctx context.Context, cocktail models.Cocktail) (models.Cocktail, error) {
	const ep = "repository.cocktail.CreateCocktail"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("TRANSACTION BEGIN ERROR in %s: %w", ep, err)
	}
	defer tx.Rollback(ctx)
	// Insert into cocktails table
	query := `INSERT INTO cocktails (name, picture, rating, description, recipe, user_id, approved) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id, creation_date, update_date`
	err = tx.QueryRow(ctx, query, cocktail.Name, cocktail.Picture, cocktail.Rating, cocktail.Description, cocktail.Recipe, cocktail.UserId, cocktail.Approved).
		Scan(&cocktail.Id, &cocktail.CreationDate, &cocktail.UpdateDate)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	fmt.Println("inserting ingred")
	// Insert ingredients
	for i := 0; i < len(cocktail.Ingredients); i++ {
		for j := 0; j < len(cocktail.Ingredients[i].Replacement); j++ {
			fmt.Println("inserting ingred in replacement ")
			err = r.insertIngredient(ctx, tx, &cocktail.Ingredients[i].Replacement[j], cocktail.Id)
			if err != nil {
				return models.Cocktail{}, fmt.Errorf("INSERT INGR REPLACEMENT ERROR in %s: %w", ep, err)
			}
			fmt.Println("check replacement id in arr:", cocktail.Ingredients[i].Replacement[j].Id)
		}
		err = r.insertIngredient(ctx, tx, &cocktail.Ingredients[i], cocktail.Id)
		if err != nil {
			return models.Cocktail{}, fmt.Errorf("INSERT INGR ERROR in %s: %w", ep, err)
		}
	}
	fmt.Println("inserting tags")
	// Insert tags
	for _, tag := range cocktail.Tags {
		err = r.insertCocktailTag(ctx, tx, cocktail.Id, tag.Id)
		if err != nil {
			return models.Cocktail{}, fmt.Errorf("INSERT TAG ERROR in %s: %w", ep, err)
		}
	}
	fmt.Println("completed")
	err = tx.Commit(ctx)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("TRANSACTION COMMIT ERROR in %s: %w", ep, err)
	}

	return cocktail, nil
}

func (r *CocktailRepository) IsCocktailExistAndApproved(ctx context.Context, name string) (bool, error) {
	const ep = "repository.cocktail.IsCocktailExistAndApproved"

	query := `SELECT EXISTS(SELECT 1 FROM cocktails WHERE name=$1 AND approved=TRUE)`
	var exists bool
	err := r.db.QueryRow(ctx, query, name).Scan(&exists)
	if err != nil {
		return exists, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
	}
	return exists, nil
}

func (r *CocktailRepository) GetCocktailsByName(ctx context.Context, name string) ([]models.Cocktail, error) {
	const ep = "repository.cocktail.GetCocktailsByName"

	query := `SELECT id, name, picture, rating, description, recipe, user_id, approved, creation_date, update_date
			  FROM cocktails WHERE name ILIKE '%' || $1 || '%'`
	rows, err := r.db.Query(ctx, query, name)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var cocktails []models.Cocktail
	for rows.Next() {
		var c models.Cocktail
		err := rows.Scan(&c.Id, &c.Name, &c.Picture, &c.Rating, &c.Description, &c.Recipe, &c.UserId, &c.Approved, &c.CreationDate, &c.UpdateDate)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}

		// Fetch ingredients and tags
		c.Ingredients, err = r.GetIngredientsByCocktailID(ctx, c.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH INGR ERROR in %s: %w", ep, err)
		}

		c.Tags, err = r.GetTagsByCocktailID(ctx, c.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH TAG ERROR in %s: %w", ep, err)
		}

		cocktails = append(cocktails, c)
	}

	return cocktails, nil
}

func (r *CocktailRepository) GetCocktails(ctx context.Context, isApproved bool, usedId string) ([]models.Cocktail, error) {
	const ep = "repository.cocktail.GetCocktails"

	query := `SELECT id, name, picture, rating, description, recipe, user_id, approved, creation_date, update_date
			  FROM cocktails WHERE (approved=true AND $1) OR (approved=false AND NOT $1 AND user_id=$2)`
	rows, err := r.db.Query(ctx, query, isApproved, usedId)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var cocktails []models.Cocktail
	for rows.Next() {
		var c models.Cocktail
		err := rows.Scan(&c.Id, &c.Name, &c.Picture, &c.Rating, &c.Description, &c.Recipe, &c.UserId, &c.Approved, &c.CreationDate, &c.UpdateDate)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}

		// Fetch ingredients and tags
		c.Ingredients, err = r.GetIngredientsByCocktailID(ctx, c.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH INGR ERROR in %s: %w", ep, err)
		}

		c.Tags, err = r.GetTagsByCocktailID(ctx, c.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH TAG ERROR in %s: %w", ep, err)
		}

		cocktails = append(cocktails, c)
	}

	return cocktails, nil
}

func (r *CocktailRepository) GetCocktailByID(ctx context.Context, id string) (models.Cocktail, error) {
	const ep = "repository.cocktail.GetCocktailByID"

	query := `SELECT id, name, picture, rating, description, recipe, user_id, approved, creation_date, update_date
			  FROM cocktails WHERE id=$1`
	var c models.Cocktail
	err := r.db.QueryRow(ctx, query, id).Scan(&c.Id, &c.Name, &c.Picture, &c.Rating, &c.Description, &c.Recipe, &c.UserId, &c.Approved, &c.CreationDate, &c.UpdateDate)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}

	// Fetch ingredients and tags
	c.Ingredients, err = r.GetIngredientsByCocktailID(ctx, c.Id)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("FETCH INGR ERROR in %s: %w", ep, err)
	}

	c.Tags, err = r.GetTagsByCocktailID(ctx, c.Id)
	if err != nil {
		return models.Cocktail{}, fmt.Errorf("FETCH TAG ERROR in %s: %w", ep, err)
	}

	return c, nil
}

func (r *CocktailRepository) GetCocktailsByIngredients(ctx context.Context, ingredientNames []string) ([]models.Cocktail, error) {
	const ep = "repository.cocktail.GetCocktailsByIngredients"

	if len(ingredientNames) == 0 {
		return nil, errors.New(ep + ": ingredientNames cannot be empty")
	}

	// Подготовка SQL-запроса
	query := `
        SELECT DISTINCT c.id, c.name, c.picture, c.rating, c.description, c.recipe, c.user_id, c.approved, c.creation_date, c.update_date
        FROM cocktails c
        INNER JOIN ingredients i ON c.id = i.cocktail_id
        INNER JOIN ingredients_nomenclature in ON i.nomenclature_id = in.id
        WHERE in.name = ANY($1)
    `

	// Выполнение запроса
	rows, err := r.db.Query(ctx, query, ingredientNames)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var cocktails []models.Cocktail
	for rows.Next() {
		var c models.Cocktail
		err := rows.Scan(&c.Id, &c.Name, &c.Picture, &c.Rating, &c.Description, &c.Recipe, &c.UserId, &c.Approved, &c.CreationDate, &c.UpdateDate)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}

		// Получение ингредиентов и тегов для каждого коктейля
		c.Ingredients, err = r.GetIngredientsByCocktailID(ctx, c.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH INGR ERROR in %s: %w", ep, err)
		}

		c.Tags, err = r.GetTagsByCocktailID(ctx, c.Id)
		if err != nil {
			return nil, fmt.Errorf("FETCH TAG ERROR in %s: %w", ep, err)
		}

		cocktails = append(cocktails, c)
	}

	return cocktails, nil
}

func (r *CocktailRepository) GetCocktailsByTags(ctx context.Context, tagNames []string) ([]models.Cocktail, error) {
	const ep = "repository.cocktail.GetCocktailsByTags"

	// Implement query to fetch cocktails by tags
	return nil, errors.New("not implemented")
}

func (r *CocktailRepository) UpdateCocktail(ctx context.Context, cocktail models.Cocktail) error {
	const ep = "repository.cocktail.UpdateCocktail"

	// Implement update functionality
	return errors.New("not implemented")
}

func (r *CocktailRepository) DeleteCocktail(ctx context.Context, id string) error {
	const ep = "repository.cocktail.DeleteCocktail"

	query := `DELETE FROM cocktails WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DELETE ERROR in %s: %w", ep, err)
	}
	return nil
}

// Additional helper methods
func (r *CocktailRepository) GetIngredientsByCocktailID(ctx context.Context, cocktailId string) ([]models.Ingredient, error) {
	const ep = "repository.cocktail.GetIngredientsByCocktailID"

	// Implement fetching ingredients by cocktail ID
	return nil, errors.New("not implemented")
}

func (r *CocktailRepository) GetTagsByCocktailID(ctx context.Context, cocktailId string) ([]models.Tag, error) {
	const ep = "repository.cocktail.GetTagsByCocktailID"

	// Implement fetching tags by cocktail ID
	return nil, errors.New("not implemented")
}

func (r *CocktailRepository) insertIngredient(ctx context.Context, tx pgx.Tx, ingredient *models.Ingredient, cocktailID string) error {
	const ep = "repository.cocktail.insertIngredient"

	fmt.Println("inserting ingredient", ingredient.Name.Name)
	query := `INSERT INTO ingredients (cocktail_id, nomenclature_id, amount, measure, optional, decorative, position) 
			  VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING id`

	fmt.Println("nomeclature id in insertingredient :", ingredient.Name.Id)
	err := tx.QueryRow(ctx, query, cocktailID, ingredient.Name.Id, ingredient.Amount, ingredient.Measure, ingredient.Optional, ingredient.Decorative, ingredient.Position).
		Scan(&ingredient.Id)
	if err != nil {
		return fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	// Insert replacements
	for i := 0; i < len(ingredient.Replacement); i++ {
		fmt.Println("ingredient id:", ingredient.Id, ingredient.Name.Name)
		fmt.Println("replacement id:", ingredient.Replacement[i].Id, ingredient.Replacement[i].Name.Name)
		err = r.insertIngredientReplacement(ctx, tx, ingredient.Id, ingredient.Replacement[i].Id)
		if err != nil {
			return fmt.Errorf("INSERT INGR REPLACEMENT ERROR in %s: %w", ep, err)
		}
	}
	return nil
}

func (r *CocktailRepository) insertIngredientReplacement(ctx context.Context, tx pgx.Tx, ingredientId, replacementId string) error {
	const ep = "repository.cocktail.insertIngredientReplacement"

	query := `INSERT INTO ingredient_replacements (ingredient_id, replacement_id) VALUES ($1, $2)`
	fmt.Println("IN insertIngredientReplacement method begin")
	_, err := tx.Exec(ctx, query, ingredientId, replacementId)
	if err != nil {
		return fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	return nil
}

func (r *CocktailRepository) insertCocktailTag(ctx context.Context, tx pgx.Tx, cocktailId, tagId string) error {
	const ep = "repository.cocktail.insertCocktailTag"

	query := `INSERT INTO cocktail_tags (cocktail_id, tag_id) VALUES ($1, $2)`
	_, err := tx.Exec(ctx, query, cocktailId, tagId)
	if err != nil {
		return fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	return nil
}
