package services

import (
	db "GoTwitter/db/repositories"
	"GoTwitter/models"
	"GoTwitter/utils"
	"context"
	"errors"
	"fmt"
)

type UserService interface {
	CreateUser(ctx context.Context, user *models.User) (*models.User, error)
	Login(ctx context.Context, email string, password string) (*models.User, string, error)
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
	if user == nil {
		return nil, errors.New("user is nil")
	}
	if user.Password == "" {
		return nil, errors.New("password is required")
	}

	hashed, err := utils.HashPassword(user.Password)
	if err != nil {
		return nil, err
	}
	user.Password = string(hashed)

	return u.userRepository.Create(ctx, user)
}

func (u *UserServiceImpl) Login(ctx context.Context, email string, password string) (*models.User, string, error) {
	if email == "" || password == "" {
		return nil, "", errors.New("email and password are required")
	}

	user, err := u.userRepository.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, "", err
	}
	if user == nil {
		return nil, "", errors.New("invalid credentials")
	}

	if !utils.CheckPasswordHash(password, user.Password) {
		return nil, "", errors.New("invalid credentials")
	}

	token, err := utils.GenerateJWT(user.Id, user.Email)
	if err != nil {
		return nil, "", err
	}
	return user, token, nil
}
