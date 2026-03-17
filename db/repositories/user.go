package db

import (
	"GoTwitter/models"
	"context"
	"database/sql"
	"errors"
)

type UserRepository interface {
	Create(ctx context.Context, user *models.User) (*models.User, error)
	GetByID(ctx context.Context, id int64) (*models.User, error)
	GetAll() ([]*models.User, error)
	DeleteByID(id int64) error
	GetUserByEmail(ctx context.Context, email string) (*models.User, error)
}

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(_db *sql.DB) UserRepository {
	return &UserRepositoryImpl{
		db: _db,
	}
}

func (u *UserRepositoryImpl) Create(ctx context.Context, user *models.User) (*models.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if user == nil {
		return nil, errors.New("user is nil")
	}
	if u.db == nil {
		return nil, errors.New("db is nil")
	}

	res, err := u.db.ExecContext(
		ctx,
		`INSERT INTO users (username, email, password, created_at, updated_at)
		 VALUES (?, ?, ?, NOW(), NOW())`,
		user.Username,
		user.Email,
		user.Password,
	)
	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return nil, err
	}

	created, err := u.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return created, nil
}

func (u *UserRepositoryImpl) GetByID(ctx context.Context, id int64) (*models.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if u.db == nil {
		return nil, errors.New("db is nil")
	}

	var user models.User
	err := u.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password, created_at, updated_at
		 FROM users
		 WHERE id = ?
		 LIMIT 1`,
		id,
	).Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (u *UserRepositoryImpl) GetAll() ([]*models.User, error) {
	if u.db == nil {
		return nil, errors.New("db is nil")
	}

	rows, err := u.db.Query(
		`SELECT id, username, email, password, created_at, updated_at
		 FROM users
		 ORDER BY id DESC`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []*models.User
	for rows.Next() {
		var user models.User
		if err := rows.Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, &user)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return users, nil
}

func (u *UserRepositoryImpl) DeleteByID(id int64) error {
	if u.db == nil {
		return errors.New("db is nil")
	}
	_, err := u.db.Exec(`DELETE FROM users WHERE id = ?`, id)
	return err
}

func (u *UserRepositoryImpl) GetUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if u.db == nil {
		return nil, errors.New("db is nil")
	}

	var user models.User
	err := u.db.QueryRowContext(
		ctx,
		`SELECT id, username, email, password, created_at, updated_at
		 FROM users
		 WHERE email = ?
		 LIMIT 1`,
		email,
	).Scan(&user.Id, &user.Username, &user.Email, &user.Password, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}
