package controller

import (
	"fmt"
	"log/slog"
	"net/http"
	"nickbar_v2/internal/models/models"
	"strings"
)

type AuthMiddleware struct {
	log         *slog.Logger
	authService authService
}

type authService interface {
	AuthUser(token string) (models.User, error)
	CheckAccess(route string, method string, user models.User) bool
}

func (AuthMiddleware *AuthMiddleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		AuthMiddleware.ProccessAuthorization(w, req, next)
	})
}

func (AuthMiddleware *AuthMiddleware) ProccessAuthorization(w http.ResponseWriter, r *http.Request, next http.Handler) {
	const ep = "controller.logging_middleware.ProccessAuthorization"

	log := AuthMiddleware.log.With(
		slog.String("ep", ep),
	)

	bearer := r.Header.Get("Authorization")
	if bearer == "" {
		log.Error("authorization header is empty")

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	bearer = strings.Replace(bearer, "Bearer ", "", -1)
	user, err := AuthMiddleware.authService.AuthUser(bearer)
	if err != nil {
		log.Error("auth user error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error auth user: %s", err)))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	isAllowed := AuthMiddleware.authService.CheckAccess(r.RequestURI, r.Method, user)
	if !isAllowed {
		log.Info("forbiden access for user", slog.Any("requestURI", r.RequestURI), slog.Any("method", r.Method), slog.Any("user", user))

		w.Write([]byte("forbidden"))
		w.WriteHeader(http.StatusForbidden)
		return
	}

	newCtx := models.WithUser(r.Context(), user)
	r = r.WithContext(newCtx)

	next.ServeHTTP(w, r)
}
