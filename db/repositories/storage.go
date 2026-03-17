package db

import "database/sql"

// Facilitates dependency injection for respository
type Storage struct {
	DB             *sql.DB
	UserRepository UserRepository
}

func NewStorage(db *sql.DB) *Storage {
	return &Storage{
		DB:             db,
		UserRepository: NewUserRepository(db),
	}
}
