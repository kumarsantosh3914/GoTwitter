package db

// Facilitates dependency injection for respository
type Storage struct {
	UserRepository UserRepository
}

func NewStorage() *Storage {
	return &Storage{
		UserRepository: &UserRepositoryImpl{},
	}
}
