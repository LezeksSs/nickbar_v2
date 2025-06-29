package service

import (
	"context"
	"errors"
	"fmt"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
)

type TagService struct {
	tagRepository TagRepository
}

type TagRepository interface {
	CreateTag(ctx context.Context, tag models.Tag) (models.Tag, error)
	IsTagExistAndApproved(ctx context.Context, name string) (bool, models.Tag, error)
	// GetTag(ctx context.Context, name string) (models.Tag, error)
	// GetTags(ctx context.Context) ([]models.Tag, error)
	GetPersonalTags(ctx context.Context, userId string) ([]models.Tag, error)
	GetUnapprovedTags(ctx context.Context) ([]models.Tag, error)
	// GetApprovedTags(ctx context.Context) ([]models.Tag, error)
	// UpdateTag(ctx context.Context, tag models.Tag) error
	UpdateTagsApprovedStatus(ctx context.Context, tags []models.Tag) error
	DeleteTag(ctx context.Context, id string) error
	GetTagsByCocktailID(ctx context.Context, cocktailId string) ([]models.Tag, error)
	InsertCocktailTag(ctx context.Context, cocktailId, tagId string) error
	DeleteCocktailTag(ctx context.Context, cocktailId, tagId, userId string, userRole int) error
}

func NewTagRepository(tagRepository TagRepository) *TagService {
	return &TagService{tagRepository: tagRepository}
}

func mapTag(userId string, req requests.TagRequest) models.Tag {
	tag := models.Tag{
		Name:   req.Name,
		UserId: userId,
	}
	return tag
}

func (ts *TagService) CreateTag(ctx context.Context, tagRequest requests.TagRequest) (models.Tag, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return models.Tag{}, errors.New("unathorized")
	}
	mappedTag := mapTag(user.Id, tagRequest)
	if user.Role == 1 {
		mappedTag.Approved = true
	}
	tag, err := ts.tagRepository.CreateTag(ctx, mappedTag)
	if err != nil {
		return models.Tag{}, fmt.Errorf("createtag in TagService error %s", err.Error())
	}
	return tag, nil
}

func (ts *TagService) AddTagToCocktail(ctx context.Context, cocktailId string, tagRequest requests.TagRequest) (models.Tag, error) {
	ok, tag, err := ts.tagRepository.IsTagExistAndApproved(ctx, tagRequest.Name)
	if err != nil {
		return models.Tag{}, fmt.Errorf("istagexist error in tagservice %s", err.Error())
	}
	if !ok {
		tag, err = ts.CreateTag(ctx, tagRequest)
		if err != nil {
			return models.Tag{}, fmt.Errorf("createtag in AddTagToCocktail error %s", err.Error())
		}
	}
	err = ts.tagRepository.InsertCocktailTag(ctx, cocktailId, tag.Id)
	if err != nil {
		return models.Tag{}, fmt.Errorf("createtag in InsertCocktailTag error %s", err.Error())
	}
	return tag, nil
}

func (ts *TagService) GetTagsByCocktailID(ctx context.Context, cocktailId string) ([]models.Tag, error) {
	tags, err := ts.tagRepository.GetTagsByCocktailID(ctx, cocktailId)
	if err != nil {
		return []models.Tag{}, fmt.Errorf("gettags in TagService error %s", err.Error())
	}
	return tags, nil
}

func (ts *TagService) GetPersonalTags(ctx context.Context) ([]models.Tag, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return []models.Tag{}, errors.New("unathorized")
	}
	tags, err := ts.tagRepository.GetPersonalTags(ctx, user.Id)
	if err != nil {
		return []models.Tag{}, fmt.Errorf("getpersonaltags in TagService error %s", err.Error())
	}
	return tags, nil
}

func (ts *TagService) GetUnapprovedTags(ctx context.Context) ([]models.Tag, error) {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return nil, errors.New("unathorized")
	}
	if user.Role == 1 {
		tags, err := ts.tagRepository.GetUnapprovedTags(ctx)
		if err != nil {
			return []models.Tag{}, fmt.Errorf("getunapprovedtags in TagService error %s", err.Error())
		}
		return tags, nil
	}
	return []models.Tag{}, errors.New("restricted access")
}

func (ts *TagService) GetApprovedTags(ctx context.Context) ([]models.Tag, error) {
	tags, err := ts.tagRepository.GetUnapprovedTags(ctx)
	if err != nil {
		return []models.Tag{}, fmt.Errorf("getapprovedtags in TagService error %s", err.Error())
	}
	return tags, nil
}

func (ts *TagService) UpdateTagsApprovedStatus(ctx context.Context, tagsRequest []requests.TagUpdateRequest) error {
	var tags []models.Tag
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	if user.Role == 1 {
		for _, req := range tagsRequest {
			tag := models.Tag{
				Id:       req.Id,
				Approved: req.Approved, // Меняем только поле Approved
			}
			tags = append(tags, tag)
		}

		err := ts.tagRepository.UpdateTagsApprovedStatus(ctx, tags)
		if err != nil {
			return fmt.Errorf("failed to update tags in TagService approved status: %w", err)
		}
		return nil
	}

	return errors.New("restricted access")
}

func (ts *TagService) DeleteTag(ctx context.Context, id string) error {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	if user.Role == 1 {
		err := ts.tagRepository.DeleteTag(ctx, id)
		if err != nil {
			return fmt.Errorf("failed to delete tag in TagService, status: %w", err)
		}
		return nil
	}

	return errors.New("restricted access")
}

func (ts *TagService) DeleteTagFromCocktail(ctx context.Context, tagId, cocktailId string) error {
	user, ok := models.GetUserFromContext(ctx)
	if !ok {
		return errors.New("unathorized")
	}

	err := ts.tagRepository.DeleteCocktailTag(ctx, cocktailId, tagId, user.Id, user.Role)
	if err != nil {
		return fmt.Errorf("failed to delete tag in TagService, status: %w", err)
	}

	return nil
}
