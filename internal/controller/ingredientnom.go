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

type IngredientNomController struct {
	log                  *slog.Logger
	ingredientNomService ingredientNomService
}

type ingredientNomService interface {
	CreateIngredientNom(ctx context.Context, nomenclature requests.IngredientNomRequest) (models.IngredientNomenclature, error)
	GetIngredientNoms(ctx context.Context) ([]models.IngredientNomenclature, error)
	GetUnapprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	GetApprovedIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	GetPersonalIngredientNom(ctx context.Context) ([]models.IngredientNomenclature, error)
	UpdateIngredientNomApprovedStatus(ctx context.Context, ingredientNomRequest []requests.IngredientNomUpdateRequest) error
	DeleteIngredientNom(ctx context.Context, id string) error
}

func NewIngredientNomController(log *slog.Logger, ingredientNomService ingredientNomService) *IngredientNomController {
	return &IngredientNomController{log: log, ingredientNomService: ingredientNomService}
}

func (inc *IngredientNomController) CreateIngredientNom(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.CreateIngredientNom"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	var req requests.IngredientNomRequest

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

	log.Info("requested body unmarshaled", slog.Any("req", req))

	nomenclature, err := inc.ingredientNomService.CreateIngredientNom(r.Context(), req)
	if err != nil {
		log.Error("creating ingredientnom error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error creating nomenclature: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(nomenclature)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (inc *IngredientNomController) GetIngredientNoms(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.GetIngredientNoms"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	nomenclatures, err := inc.ingredientNomService.GetIngredientNoms(r.Context())
	if err != nil {
		log.Error("getting ingredientnom error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting nomenclatures: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(nomenclatures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (inc *IngredientNomController) GetPersonalIngredientNom(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.GetPersonalIngredientNom"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	nomenclatures, err := inc.ingredientNomService.GetPersonalIngredientNom(r.Context())
	if err != nil {
		log.Error("getting personal ingredientnom error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting personal nomenclatures: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(nomenclatures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (inc *IngredientNomController) GetUnapprovedIngredientNom(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.GetUnapprovedIngredientNom"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	nomenclatures, err := inc.ingredientNomService.GetUnapprovedIngredientNom(r.Context())
	if err != nil {
		log.Error("getting unapproved ingredientnom error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting unapproved nomenclatures: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(nomenclatures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (inc *IngredientNomController) GetApprovedIngredientNom(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.GetApprovedIngredientNom"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	nomenclatures, err := inc.ingredientNomService.GetApprovedIngredientNom(r.Context())
	if err != nil {
		log.Error("getting approved ingredientnom error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting approved nomenclatures: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(nomenclatures)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (inc *IngredientNomController) UpdateIngredientNomApprovedStatus(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.UpdateIngredientNomApprovedStatus"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	var req []requests.IngredientNomUpdateRequest

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

	log.Info("requested body unmarshaled", slog.Any("req", req))

	err = inc.ingredientNomService.UpdateIngredientNomApprovedStatus(r.Context(), req)
	if err != nil {
		log.Error("updating ingredientnom approval status error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error updating nomenclature: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (inc *IngredientNomController) DeleteIngredientNom(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.ingredientnom.DeleteIngredientNom"

	log := inc.log.With(
		slog.String("ep", ep),
	)

	var req requests.DeleteEntityRequest

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

	log.Info("requested body unmarshaled", slog.Any("req", req))

	err = inc.ingredientNomService.DeleteIngredientNom(r.Context(), req.Id)
	if err != nil {
		log.Error("deleting ingredientnom error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error deleting nomeclature: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
