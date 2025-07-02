package controller

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"nickbar_v2/internal/models/requests"
	"nickbar_v2/internal/repository"
)

type AuthController struct {
	log         *slog.Logger
	userService userService
}

type userService interface {
	LoginUser(nickname requests.LoginRequest) (repository.TokenPayload, error)
	RegisterUser(nickname requests.LoginRequest) (repository.TokenPayload, error)
}

func NewAuthController(log *slog.Logger, userService userService) *AuthController {
	return &AuthController{log: log, userService: userService}
}

func (as *AuthController) LoginUser(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.auth.LoginUser"

	var req requests.LoginRequest

	log := as.log.With(
		slog.String("ep", ep),
	)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("request body is empty")

		w.Write([]byte("Bad gateway"))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = json.Unmarshal(payload, &req)
	if err != nil {
		log.Error("failed to unmarshal request body", slog.Any("err", err))

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("requested body unmarshaled", slog.Any("request", req))

	token, err := as.userService.LoginUser(req)
	if err != nil {
		log.Error("login user error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error login user: %s", err)))
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	resp, _ := json.Marshal(token)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (as *AuthController) RegisterUser(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.auth.RegisterUser"

	var req requests.LoginRequest

	log := as.log.With(
		slog.String("ep", ep),
	)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("request body is empty")

		w.Write([]byte("Bad gateway"))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = json.Unmarshal(payload, &req)
	if err != nil {
		log.Error("failed to unmarshal request body", slog.Any("err", err))

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("requested body unmarshaled", slog.Any("request", req))

	token, err := as.userService.RegisterUser(req)
	if err != nil {
		log.Error("register user error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error register user: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(token)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}
