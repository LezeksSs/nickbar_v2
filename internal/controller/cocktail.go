package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
)

type CocktailController struct {
	log             *slog.Logger
	cocktailService cocktailService
}
type cocktailService interface {
	CreateCocktail(ctx context.Context, data requests.CocktailRequest) (models.Cocktail, error)
	GetCocktails(ctx context.Context, isApproved bool) ([]models.Cocktail, error)
	// UpdateCocktail(ctx context.Context, data requests.CocktailRequest) error
	// GetCocktailsByName(ctx context.Context, name string) ([]models.Cocktail, error)
	SearchCocktails(ctx context.Context, name string, ingredients []requests.IngredientRequest, tags []requests.TagRequest, pagination models.Pagination) ([]models.Cocktail, error)
	DeleteCocktail(ctx context.Context, id string) error
}

func NewCocktailController(log *slog.Logger, cocktailService cocktailService) *CocktailController {
	return &CocktailController{log: log, cocktailService: cocktailService}
}

func (cc *CocktailController) CreateCocktail(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.cocktail.CreateCocktail"

	var req requests.CocktailRequest

	log := cc.log.With(
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

	cocktail, err := cc.cocktailService.CreateCocktail(r.Context(), req)
	if err != nil {
		log.Error("cocktail creating error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting cocktail: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(cocktail)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (cc *CocktailController) GetCocktails(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.cocktail.GetCocktails"

	type RequestBody struct {
		IsApproved bool `json:"IsApproved"`
	}

	var req RequestBody

	log := cc.log.With(
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

	cocktails, err := cc.cocktailService.GetCocktails(r.Context(), req.IsApproved)
	if err != nil {
		log.Error("cocktail creating error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting cocktail: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(cocktails)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (cc *CocktailController) SearchCocktails(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.cocktail.SearchCocktails"

	var req requests.SearchRequest

	log := cc.log.With(
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

	cocktail, err := cc.cocktailService.SearchCocktails(r.Context(), req.Name, req.Ingredients, req.Tags, req.Pagination)
	if err != nil {
		log.Error("cocktail searching error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error searching cocktail: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(cocktail)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (cc *CocktailController) DeleteCocktail(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.cocktail.DeleteCocktail"

	var req requests.DeleteEntityRequest

	log := cc.log.With(
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

	err = cc.cocktailService.DeleteCocktail(r.Context(), req.Id)
	if err != nil {
		log.Error("cocktail deleting error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error deleting cocktail: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
