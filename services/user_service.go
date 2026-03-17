package services

import (
	db "GoTwitter/db/repositories"
	"GoTwitter/models"
	"context"
	"fmt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
}

type UserServiceImpl struct {
	userRepository db.UserRepository
}

func NewUserService(_userRepository db.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: _userRepository,
	}
}

func (u *UserServiceImpl) CreateUser(ctx context.Context, user *models.User) (*models.User, error) {
	fmt.Println("Creating user in userService")
	return u.userRepository.Create(ctx, user)
}
