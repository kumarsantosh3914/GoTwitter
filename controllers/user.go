package controllers

import (
	"GoTwitter/config/env"
	"GoTwitter/models"
	"GoTwitter/services"
	"GoTwitter/utils"
	"log"
	"net/http"
	"strings"
	"time"
)

type UserController struct {
	UserService services.UserService
}

func NewUserController(_userService services.UserService) *UserController {
	return &UserController{
		UserService: _userService,
	}
}

type userResponse struct {
	Id        int64     `json:"id"`
	Username  string    `json:"username"`
	Email     string    `json:"email"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func toUserResponse(u *models.User) *userResponse {
	if u == nil {
		return nil
	}
	return &userResponse{
		Id:        u.Id,
		Username:  u.Username,
		Email:     u.Email,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

func setAuthCookie(w http.ResponseWriter, token string) {
	secure := strings.EqualFold(env.GetString("COOKIE_SECURE", "false"), "true")
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    token,
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	secure := strings.EqualFold(env.GetString("COOKIE_SECURE", "false"), "true")
	http.SetCookie(w, &http.Cookie{
		Name:     "auth_token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
		MaxAge:   -1,
	})
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	created, err := uc.UserService.CreateUser(r.Context(), &models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		log.Printf("[ERROR] signup create user failed: %v", err)
		utils.WriteJsonErrorResponse(w, http.StatusInternalServerError, "failed to create user", err)
		return
	}

	// Issue auth cookie on signup
	user, token, err := uc.UserService.Login(r.Context(), payload.Email, payload.Password)
	if err == nil && token != "" {
		setAuthCookie(w, token)
		created = user
	} else if err != nil {
		log.Printf("[WARN] signup auto-login failed: %v", err)
	}

	utils.WriteJsonResponse(w, http.StatusCreated, toUserResponse(created))
}

func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "invalid json body", err)
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		utils.WriteJsonErrorResponse(w, http.StatusBadRequest, "validation failed", err)
		return
	}

	user, token, err := uc.UserService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		log.Printf("[WARN] login failed: %v", err)
		utils.WriteJsonErrorResponse(w, http.StatusUnauthorized, "invalid credentials", err)
		return
	}

	setAuthCookie(w, token)
	utils.WriteJsonResponse(w, http.StatusOK, toUserResponse(user))
}

func (uc *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
