package repository

import (
	"context"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IngredientNomenclatureRepository struct {
	db *pgxpool.Pool
}

func NewIngredientNomenclatureRepository(db *pgxpool.Pool) *IngredientNomenclatureRepository {
	return &IngredientNomenclatureRepository{db: db}
}

func (r *IngredientNomenclatureRepository) CreateIngredientNom(ctx context.Context, nomenclature models.IngredientNomenclature) (models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.CreateIngredientNom"

	query := `INSERT INTO ingredients_nomenclature (name, picture, user_id, approved)
			  VALUES ($1, $2, $3, $4) RETURNING id`
	err := r.db.QueryRow(ctx, query, nomenclature.Name, nomenclature.Picture, nomenclature.UserId, nomenclature.Approved).
		Scan(&nomenclature.Id)
	if err != nil {
		return models.IngredientNomenclature{}, fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	return nomenclature, nil
}

func (r *IngredientNomenclatureRepository) GetIngredientNom(ctx context.Context, name string) (models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.GetIngredientNom"

	query := `SELECT id, name, picture, user_id, approved FROM ingredients_nomenclature WHERE name=$1`
	var nom models.IngredientNomenclature
	err := r.db.QueryRow(ctx, query, name).Scan(&nom.Id, &nom.Name, &nom.Picture, &nom.UserId, &nom.Approved)
	if err != nil {
		return models.IngredientNomenclature{}, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	return nom, err
}

func (r *IngredientNomenclatureRepository) GetIngredientNoms(ctx context.Context) ([]models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.GetIngredientNoms"

	query := `SELECT id, name, picture, user_id, approved FROM ingredients_nomenclature`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var noms []models.IngredientNomenclature
	for rows.Next() {
		var nom models.IngredientNomenclature
		err := rows.Scan(&nom.Id, &nom.Name, &nom.Picture, &nom.UserId, &nom.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		noms = append(noms, nom)
	}

	return noms, nil
}

func (r *IngredientNomenclatureRepository) GetUnapprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.GetUnapprovedIngredientNom"

	query := `SELECT id, name, picture, user_id, approved FROM ingredients_nomenclature WHERE approved=FALSE`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var noms []models.IngredientNomenclature
	for rows.Next() {
		var nom models.IngredientNomenclature
		err := rows.Scan(&nom.Id, &nom.Name, &nom.Picture, &nom.UserId, &nom.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		noms = append(noms, nom)
	}

	return noms, nil
}

func (r *IngredientNomenclatureRepository) GetApprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.GetApprovedIngredientNom"

	query := `SELECT id, name, picture, user_id, approved FROM ingredients_nomenclature WHERE approved=TRUE`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var noms []models.IngredientNomenclature
	for rows.Next() {
		var nom models.IngredientNomenclature
		err := rows.Scan(&nom.Id, &nom.Name, &nom.Picture, &nom.UserId, &nom.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		noms = append(noms, nom)
	}

	return noms, nil
}

func (r *IngredientNomenclatureRepository) GetPersonalIngredientNom(ctx context.Context, userId string) ([]models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.GetPersonalIngredientNom"

	query := `SELECT id, name, picture, user_id, approved FROM ingredients_nomenclature WHERE user_id=$1`
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var noms []models.IngredientNomenclature
	for rows.Next() {
		var nom models.IngredientNomenclature
		err := rows.Scan(&nom.Id, &nom.Name, &nom.Picture, &nom.UserId, &nom.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
		}
		noms = append(noms, nom)
	}

	return noms, nil
}

func (r *IngredientNomenclatureRepository) UpdateIngredientNom(ctx context.Context, nomenclature models.IngredientNomenclature) error {
	const ep = "repository.ingredient_nom.UpdateIngredientNom"

	query := `UPDATE ingredients_nomenclature SET name=$1, picture=$2, approved=$3 WHERE id=$4`
	_, err := r.db.Exec(ctx, query, nomenclature.Name, nomenclature.Picture, nomenclature.Approved, nomenclature.Id)
	if err != nil {
		return fmt.Errorf("UPDATE ERROR in %s: %w", ep, err)
	}
	return nil
}

func (r *IngredientNomenclatureRepository) UpdateIngredientNomsApprovedStatus(ctx context.Context, nomenclatures []models.IngredientNomenclature) error {
	const ep = "repository.ingredient_nom.UpdateIngredientNomsApprovedStatus"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TRANSACTION BEGIN ERROR in %s: %w", ep, err)
	}
	defer tx.Rollback(ctx)

	for _, nom := range nomenclatures {
		query := `UPDATE ingredients_nomenclature SET approved=$1 WHERE id=$2`
		_, err := tx.Exec(ctx, query, nom.Approved, nom.Id)
		if err != nil {
			return fmt.Errorf("UPDATE ERROR in %s: %w", ep, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("TRANSACTION COMMIT ERROR in %s: %w", ep, err)
	}
	return nil
}

func (r *IngredientNomenclatureRepository) DeleteIngredientNom(ctx context.Context, id string) error {
	const ep = "repository.ingredient_nom.DeleteIngredientNom"

	query := `DELETE FROM ingredients_nomenclature WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DELETE ERROR in %s: %w", ep, err)
	}
	return nil
}

func (r *IngredientNomenclatureRepository) IsIngredientNomExistAndApproved(ctx context.Context, name string) (bool, models.IngredientNomenclature, error) {
	const ep = "repository.ingredient_nom.IsIngredientNomExistAndApproved"

	query := `SELECT id, name, picture, user_id, approved FROM ingredients_nomenclature WHERE name=$1 AND (approved=TRUE OR user_id=$2)`
	var ingredient models.IngredientNomenclature
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return false, models.IngredientNomenclature{}, errors.New("unathorized")
	}
	err := r.db.QueryRow(ctx, query, name, user.Id).Scan(
		&ingredient.Id,
		&ingredient.Name,
		&ingredient.Picture,
		&ingredient.UserId,
		&ingredient.Approved,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, models.IngredientNomenclature{}, nil
		}
		return false, models.IngredientNomenclature{}, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
	}

	return true, ingredient, nil
}
