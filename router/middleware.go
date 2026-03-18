package router

import (
	apperrors "GoTwitter/errors"
	"GoTwitter/utils"
	"context"
	"net/http"
)

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie("auth_token")
		if err != nil {
			apperrors.WriteError(w, apperrors.NewAppError("unauthorized: missing auth token", http.StatusUnauthorized, err))
			return
		}

		claims, err := utils.ParseJWT(cookie.Value)
		if err != nil {
			apperrors.WriteError(w, apperrors.NewAppError("unauthorized: invalid token", http.StatusUnauthorized, err))
			return
		}

		ctx := context.WithValue(r.Context(), utils.UserContextKey, claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
