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

type TagController struct {
	log        *slog.Logger
	tagService tagService
}

type tagService interface {
	CreateTag(ctx context.Context, tagRequest requests.TagRequest) (models.Tag, error)
	GetTagsByCocktailID(ctx context.Context, cocktailId string) ([]models.Tag, error)
	GetPersonalTags(ctx context.Context) ([]models.Tag, error)
	GetUnapprovedTags(ctx context.Context) ([]models.Tag, error)
	GetApprovedTags(ctx context.Context) ([]models.Tag, error)
	UpdateTagsApprovedStatus(ctx context.Context, tagsRequest []requests.TagUpdateRequest) error
	DeleteTag(ctx context.Context, id string) error
}

func NewTagController(log *slog.Logger, tagService tagService) *TagController {
	return &TagController{log: log, tagService: tagService}
}

func (tc *TagController) CreateTag(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.CreateTag"

	log := tc.log.With(
		slog.String("ep", ep),
	)

	var req requests.TagRequest

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

	tag, err := tc.tagService.CreateTag(r.Context(), req)
	if err != nil {
		log.Error("creating tag error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error creating tag: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(tag)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (tc *TagController) GetTags(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.GetTags"

	log := tc.log.With(
		slog.String("ep", ep),
	)

	var cocktailId string

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		log.Error("request body is empty")

		w.Write([]byte("Bad gateway"))
		w.WriteHeader(http.StatusBadGateway)
		return
	}

	err = json.Unmarshal(payload, &cocktailId)
	if err != nil {
		log.Error("failed to unmarshal request body", slog.Any("err", err))

		w.Write([]byte("Bad request"))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	log.Info("requested body unmarshaled", slog.Any("req", cocktailId))

	tags, err := tc.tagService.GetTagsByCocktailID(r.Context(), cocktailId)
	if err != nil {
		log.Error("creating tag error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting  tags: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(tags)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (tc *TagController) GetPersonalTags(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.GetPersonalTags"

	log := tc.log.With(
		slog.String("ep", ep),
	)

	tags, err := tc.tagService.GetPersonalTags(r.Context())
	if err != nil {
		log.Error("getting personal tags error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting personal tags: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(tags)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (tc *TagController) GetUnapprovedTags(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.GetUnapprovedTags"

	log := tc.log.With(
		slog.String("ep", ep),
	)

	tags, err := tc.tagService.GetUnapprovedTags(r.Context())
	if err != nil {
		log.Error("getting unapproved tags error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting personal tags: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(tags)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (tc *TagController) GetApprovedTags(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.GetApprovedTags"

	log := tc.log.With(
		slog.String("ep", ep),
	)

	tags, err := tc.tagService.GetApprovedTags(r.Context())
	if err != nil {
		log.Error("getting approved tags error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting approved tags: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	resp, _ := json.Marshal(tags)
	w.Header().Set("Content-Type", "application/json")
	w.Write(resp)
}

func (tc *TagController) UpdateTagsApprovedStatus(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.UpdateTagsApprovedStatus"

	log := tc.log.With(
		slog.String("ep", ep),
	)

	var req []requests.TagUpdateRequest

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

	err = tc.tagService.UpdateTagsApprovedStatus(r.Context(), req)
	if err != nil {
		log.Error("updating tags approved status error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error getting tag: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (tc *TagController) DeleteTag(w http.ResponseWriter, r *http.Request) {
	const ep = "controller.tag.DeleteTag"

	log := tc.log.With(
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

	err = tc.tagService.DeleteTag(r.Context(), req.Id)
	if err != nil {
		log.Error("deleting tag error", slog.Any("err", err))

		w.Write([]byte(fmt.Sprintf("Error deleting tag: %s", err)))
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}
