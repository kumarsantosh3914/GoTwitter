package controllers

import (
	"GoTwitter/services"
	"GoTwitter/models"
	"encoding/json"
	"fmt"
	"net/http"
)

type UserController struct {
	UserService services.UserService
}

func NewUserController(_userService services.UserService) *UserController {
	return &UserController{
		UserService: _userService,
	}
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Registeruser called in UserController")
	var payload struct {
		Username string `json:"username"`
		Email    string `json:"email"`
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
		http.Error(w, "invalid json body", http.StatusBadRequest)
		return
	}

	created, err := uc.UserService.CreateUser(r.Context(), &models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		http.Error(w, "failed to create user", http.StatusInternalServerError)
		return
	}
	if created == nil {
		http.Error(w, "user not created", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(created)
}
