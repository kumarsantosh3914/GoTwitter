package controllers

import (
	apperrors "GoTwitter/errors"
	"GoTwitter/config/env"
	"GoTwitter/models"
	"GoTwitter/services"
	"GoTwitter/utils"
	"errors"
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

func handleError(w http.ResponseWriter, err error) {
	var appErr *apperrors.AppError
	if errors.As(err, &appErr) {
		apperrors.WriteError(w, appErr)
		return
	}
	apperrors.WriteError(w, apperrors.NewAppError("internal server error", http.StatusInternalServerError, err))
}

func (uc *UserController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Username string `json:"username" validate:"required,min=3,max=30"`
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=8"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	created, err := uc.UserService.CreateUser(r.Context(), &models.User{
		Username: payload.Username,
		Email:    payload.Email,
		Password: payload.Password,
	})
	if err != nil {
		log.Printf("[ERROR] signup create user failed: %v", err)
		handleError(w, err)
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

	utils.WriteJsonSuccessResponse(w, http.StatusCreated, "User registered successfully", toUserResponse(created))
}

func (uc *UserController) Login(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := utils.ReadJsonBody(r, &payload); err != nil {
		handleError(w, apperrors.NewAppError("invalid json body", http.StatusBadRequest, err))
		return
	}

	if err := utils.Validator.Struct(payload); err != nil {
		handleError(w, apperrors.NewAppError("validation failed", http.StatusBadRequest, err))
		return
	}

	user, token, err := uc.UserService.Login(r.Context(), payload.Email, payload.Password)
	if err != nil {
		log.Printf("[WARN] login failed: %v", err)
		handleError(w, err)
		return
	}

	setAuthCookie(w, token)
	utils.WriteJsonSuccessResponse(w, http.StatusOK, "Login successful", toUserResponse(user))
}

func (uc *UserController) Logout(w http.ResponseWriter, r *http.Request) {
	clearAuthCookie(w)
	w.WriteHeader(http.StatusNoContent)
}
