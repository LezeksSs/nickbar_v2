package repository

import (
	"context"
	"database/sql"
	"fmt"
	"nickbar_v2/internal/models/models"

	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	db *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: pool}
}

func (r *UserRepository) GetUser(nickname string) (models.User, error) {
	const ep = "repository.user.GetUser"
	if r == nil {
		return models.User{}, fmt.Errorf("no connection to the database")
	}

	user := models.User{}
	res := r.db.QueryRow(context.Background(), "SELECT id, nickname, register_date, last_login_date, picture, role FROM users WHERE nickname = $1", nickname)
	err := res.Scan(&user.Id, &user.Nickname, &user.Register_date, &user.Last_login_date, &user.Picture, &user.Role)
	if err != nil {
		return models.User{}, fmt.Errorf("SELECT ERROR in %s: %w", ep, err.Error())
	}

	return user, nil
}

func (r *UserRepository) CreateUser(nickname string) (models.User, error) {
	const ep = "repository.user.CreateUser"

	if r == nil {
		return models.User{}, fmt.Errorf("no connection to the database")
	}
	_, err := r.db.Exec(context.Background(), "INSERT INTO users (nickname) VALUES ($1)", nickname)
	if err != nil {
		return models.User{}, fmt.Errorf("INSERT ERROR in %s: %w", ep, err)
	}

	return r.GetUser(nickname)
}

func (r *UserRepository) UpdatePicture(user models.User) error {
	const ep = "repository.user.UpdatePicture"

	if r == nil {
		return fmt.Errorf("no connection to the database")
	}

	_, err := r.db.Exec(context.Background(), `UPDATE users SET picture = $1 WHERE id = $2`, user.Picture, user.Id)

	if err != nil {
		return fmt.Errorf("UPDATE ERROR in %s: %w", ep, err)
	}

	return nil
}

func (r *UserRepository) UpdateRole(user models.User) error {
	const ep = "repository.user.UpdateRole"

	if r == nil {
		return fmt.Errorf("no connection to the database")
	}

	_, err := r.db.Exec(context.Background(), `UPDATE users SET role = $1 WHERE id = $2`, user.Role, user.Id)

	if err != nil {
		return fmt.Errorf("UPDATE ERROR in %s: %w", ep, err)
	}

	return nil
}

func (r *UserRepository) DeleteUser(nickname string) error {
	const ep = "repository.user.DeleteUser"

	if r == nil {
		return fmt.Errorf("no connection to the database")
	}

	_, err := r.db.Exec(context.Background(), `DELETE FROM users WHERE nickname = $1`, nickname)

	if err != nil {
		return fmt.Errorf("DELETE ERROR in %s: %w", ep, err)
	}

	return nil
}

func (r *UserRepository) GetUserList() ([]models.User, error) {
	const ep = "repository.user.GetUserList"

	if r == nil {
		return nil, fmt.Errorf("no connection to the database")
	}

	rows, err := r.db.Query(context.Background(), `SELECT id, nickname, register_date, last_login_date, picture, role FROM users`)

	if err != nil {
		return nil, fmt.Errorf("SELECT ERROR in %s: %w", ep, err)
	}

	defer rows.Close()

	var users []models.User

	if err = rows.Err(); err != nil {
		return nil, fmt.Errorf("ROWS READING ERROR in %s: %w", ep, err)
	}

	for rows.Next() {
		var user models.User
		var nullPic sql.NullString
		err := rows.Scan(&user.Id, &user.Nickname, &user.Register_date, &user.Last_login_date, &nullPic, &user.Role)

		if err != nil {
			return nil, fmt.Errorf("SCAN ERROR in %s: %w", ep, err)
		}

		// Присваиваем значение из nullPic в поле Picture структуры User
		if nullPic.Valid {
			user.Picture = nullPic.String // Если значение не NULL, присваиваем его
		} else {
			user.Picture = "" // Если значение NULL, присваиваем ""
		}

		users = append(users, user)
	}

	return users, nil
}
