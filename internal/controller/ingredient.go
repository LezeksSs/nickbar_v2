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

type IngredientController struct {
	log               *slog.Logger
	ingredientService ingredientService
}

type ingredientService interface {
	AddIngredientToCocktail(ctx context.Context, cocktailId string, request requests.IngredientRequest) (models.Ingredient, error)
	GetIngredientsByCocktailID(ctx context.Context, cocktailId string) ([]models.Ingredient, error)
	DeleteIngredientFromCocktail(ctx context.Context, cocktailId, ingredientId string) error
}

func NewIngredientController(log *slog.Logger, ingredientService ingredientService) *IngredientController {
	return &IngredientController{log: log, ingredientService: ingredientService}
}

func (ic *IngredientController) AddIngredientToCocktail(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredient.AddIngredientToCocktail"

	type RequestBody struct {
		CocktailID string                     `json:"cocktail_id"`
		Req        requests.IngredientRequest `json:"req"`
	}

	var requestBody RequestBody

	log := ic.log.With(
		slog.String("ep", ep),
	)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("request body is empty")

		w.Write([]byte("Bad gateway"))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = json.Unmarshal(payload, &requestBody)
	if err != nil {
		log.Error("failed to unmarshal request body", slog.Any("err", err))

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("requested body unmarshaled", slog.Any("request", requestBody))

	ingredient, err := ic.ingredientService.AddIngredientToCocktail(r.Context(), requestBody.CocktailID, requestBody.Req)
	if err != nil {
		log.Error("adding ingredient to cocktail error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error adding ingredient to cocktail: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(ingredient)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (ic *IngredientController) GetIngredients(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredient.GetIngredients"

	log := ic.log.With(
		slog.String("ep", ep),
	)

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("request body is empty")

		w.Write([]byte("Bad gateway"))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	var cocktailId string
	err = json.Unmarshal(payload, &cocktailId)
	if err != nil {
		log.Error("failed to unmarshal request body", slog.Any("err", err))

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("requested body unmarshaled", slog.Any("cocktailId", cocktailId))

	ingredients, err := ic.ingredientService.GetIngredientsByCocktailID(r.Context(), cocktailId)
	if err != nil {
		log.Error("getting ingredient error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting ingredients: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(ingredients)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (ic *IngredientController) DeleteIngredientFromCocktail(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredient.DeleteIngredientFromCocktail"

	log := ic.log.With(
		slog.String("ep", ep),
	)

	type RequestBody struct {
		CocktailID string                       `json:"cocktail_id"`
		Req        requests.DeleteEntityRequest `json:"req"`
	}

	var requestBody RequestBody

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("request body is empty")

		w.Write([]byte("Bad gateway"))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = json.Unmarshal(payload, &requestBody)
	if err != nil {
		log.Error("failed to unmarshal request body", slog.Any("err", err))

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("requested body unmarshaled", slog.Any("request", requestBody))

	err = ic.ingredientService.DeleteIngredientFromCocktail(r.Context(), requestBody.CocktailID, requestBody.Req.Id)
	if err != nil {
		log.Error("deleting ingredient from cocktail error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error deleting ingredient: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
