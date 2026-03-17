package db

import (
	"database/sql"
	"fmt"
)

type UserRepository interface {
	Create() error
}

type UserRepositoryImpl struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) UserRepository {
	return &UserRepositoryImpl{
		db: db,
	}
}

func (u *UserRepositoryImpl) Create() error {
	fmt.Println("Creating user in UserRepository")
	return nil
}
