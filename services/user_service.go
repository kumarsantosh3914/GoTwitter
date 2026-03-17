package services

import (
	db "GoTwitter/db/repositories"
	"fmt"
)

type UserService interface {
	CreateUser() error
}

type UserServiceImpl struct {
	userRepository db.UserRepository
}

func NewUserService(_userRepository db.UserRepository) UserService {
	return &UserServiceImpl{
		userRepository: _userRepository,
	}
}

func (u *UserServiceImpl) CreateUser() error {
	fmt.Println("Creating user in userService")
	u.userRepository.Create()
	return nil
}
