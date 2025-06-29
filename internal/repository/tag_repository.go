package repository

import (
	"context"
	"fmt"
	"nickbar_v2/internal/models/models"

	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(pool *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: pool}
}

func (r *TagRepository) CreateTag(ctx context.Context, tag models.Tag) (models.Tag, error) {
	const ep = "repository.tag.CreateTag"
	query := `INSERT INTO tags (name, user_id, approved) VALUES ($1, $2, $3) RETURNING id`
	err := r.db.QueryRow(ctx, query, tag.Name, tag.UserId, tag.Approved).Scan(&tag.Id)
	if err != nil {
		return models.Tag{}, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
	}
	return tag, nil
}

func (r *TagRepository) IsTagExistAndApproved(ctx context.Context, name string) (bool, models.Tag, error) {
	const ep = "repository.tag.IsTagExistAndApproved"

	query := `SELECT id, name, user_id, approved FROM tags WHERE name=$1 AND (approved=TRUE OR user_id=$2)`
	var tag models.Tag

	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return false, models.Tag{}, fmt.Errorf("unathorized")
	}

	err := r.db.QueryRow(ctx, query, name, user.Id).Scan(
		&tag.Id,
		&tag.Name,
		&tag.UserId,
		&tag.Approved,
	)

	if err != nil {
		if err == pgx.ErrNoRows {
			return false, models.Tag{}, fmt.Errorf("NOROWS ERROR in %s: %w", ep, err)
		}
		return false, models.Tag{}, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
	}

	return true, tag, nil
}

func (r *TagRepository) GetTag(ctx context.Context, name string) (models.Tag, error) {
	const ep = "repository.tag.GetTag"

	query := `SELECT id, name, user_id, approved FROM tags WHERE name=$1`
	var tag models.Tag
	err := r.db.QueryRow(ctx, query, name).Scan(&tag.Id, &tag.Name, &tag.UserId, &tag.Approved)
	if err != nil {
		return models.Tag{}, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
	}
	return tag, nil
}

func (r *TagRepository) GetTags(ctx context.Context) ([]models.Tag, error) {
	const ep = "repository.tag.GetTags"

	query := `SELECT id, name, user_id, approved FROM tags`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.Id, &tag.Name, &tag.UserId, &tag.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) GetPersonalTags(ctx context.Context, userId string) ([]models.Tag, error) {
	const ep = "repository.tag.GetPersonalTags"

	query := `SELECT id, name, user_id, approved FROM tags WHERE user_id=$1`
	rows, err := r.db.Query(ctx, query, userId)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.Id, &tag.Name, &tag.UserId, &tag.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) GetUnapprovedTags(ctx context.Context) ([]models.Tag, error) {
	const ep = "repository.tag.GetUnapprovedTags"

	query := `SELECT id, name, user_id, approved FROM tags WHERE approved=FALSE`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.Id, &tag.Name, &tag.UserId, &tag.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) GetApprovedTags(ctx context.Context) ([]models.Tag, error) {
	const ep = "repository.tag.GetApprovedTags"

	query := `SELECT id, name, user_id, approved FROM tags WHERE approved=TRUE`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.Id, &tag.Name, &tag.UserId, &tag.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) UpdateTag(ctx context.Context, tag models.Tag) error {
	const ep = "repository.tag.UpdateTag"

	query := `UPDATE tags SET name=$1, approved=$2 WHERE id=$3`
	_, err := r.db.Exec(ctx, query, tag.Name, tag.Approved, tag.Id)
	if err != nil {
		return fmt.Errorf("UPDATE ERROR in %s: %w", ep, err)
	}
	return nil
}

func (r *TagRepository) UpdateTagsApprovedStatus(ctx context.Context, tags []models.Tag) error {
	const ep = "repository.tag.UpdateTagsApprovedStatus"

	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("TRANSACTION BEGIN ERROR in %s: %w", ep, err)
	}
	defer tx.Rollback(ctx)

	for _, tag := range tags {
		query := `UPDATE tags SET approved=$1 WHERE id=$2`
		_, err := tx.Exec(ctx, query, tag.Approved, tag.Id)
		if err != nil {
			return fmt.Errorf("UPDATE ERROR in %s: %w", ep, err)
		}
	}

	err = tx.Commit(ctx)
	if err != nil {
		return fmt.Errorf("TRANSACTION COMMIT ERROR in %s: %w", ep, err)
	}
	return err
}

func (r *TagRepository) DeleteTag(ctx context.Context, id string) error {
	const ep = "repository.tag.DeleteTag"

	query := `DELETE FROM tags WHERE id=$1`
	_, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return fmt.Errorf("DELETE ERROR in %s: %w", ep, err)
	}
	return err
}

func (r *TagRepository) GetTagsByCocktailID(ctx context.Context, cocktailId string) ([]models.Tag, error) {
	const ep = "repository.tag.GetTagsByCocktailID"

	query := `SELECT t.id, t.name, t.user_id, t.approved
			  FROM tags t
			  INNER JOIN cocktail_tags ct ON t.id = ct.tag_id
			  WHERE ct.cocktail_id=$1`
	rows, err := r.db.Query(ctx, query, cocktailId)
	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		err := rows.Scan(&tag.Id, &tag.Name, &tag.UserId, &tag.Approved)
		if err != nil {
			return nil, fmt.Errorf("SCAN ROW ERROR in %s: %w", ep, err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *TagRepository) InsertCocktailTag(ctx context.Context, cocktailId, tagId string) error {
	const ep = "repository.tag.InsertCocktailTag"

	query := `INSERT INTO cocktail_tags (cocktail_id, tag_id) VALUES ($1, $2)`
	_, err := r.db.Query(ctx, query, cocktailId, tagId)
	if err != nil {
		return fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}
	return nil
}

func (r *TagRepository) DeleteCocktailTag(ctx context.Context, cocktailId, tagId, userId string, userRole int) error {
	const ep = "repository.tag.DeleteCocktailTag"
	// Получение пользователя из контекста

	// SQL-запрос с JOIN для проверки прав и удаления
	query := `
		DELETE FROM cocktail_tags
		WHERE cocktail_id = $1
		  AND tag_id = $2
		  AND EXISTS (
		      SELECT 1
		      FROM cocktails
		      WHERE cocktails.id = cocktail_tags.cocktail_id
		        AND (cocktails.user_id = $3 OR $4 = 1) -- Проверка владельца или роли администратора
		  )
	`

	// Выполнение запроса
	result, err := r.db.Exec(ctx, query, cocktailId, tagId, userId, userRole)
	if err != nil {
		return fmt.Errorf("DELETE ERROR in %s: %w", ep, err)
	}

	// Проверка количества удалённых строк
	rowsAffected := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("%s: No such tag-cocktail link exists or insufficient permissions", ep)
	}

	return nil
}
