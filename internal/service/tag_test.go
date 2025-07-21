package service

import (
	"context"
	"errors"
	"nickbar_v2/internal/mocks"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// вспомогалка создаёт сервис с мок-репозиторием
func newTagService(t *testing.T) (*TagService, *mocks.TagRepository) {
	repo := mocks.NewTagRepository(t)
	return NewTagService(repo), repo
}

func TestTagService_CreateTag(t *testing.T) {
	tests := []struct {
		name      string
		user      models.User
		req       requests.TagRequest
		saved     models.Tag // то, что «репо» вернёт
		wantErr   bool
		wantTag   models.Tag // ожидаемый результат сервиса
		setupMock func(*mocks.TagRepository, context.Context, models.Tag)
	}{
		{
			name:    "admin -> tag approved",
			user:    models.User{Id: "u1", Role: 1},
			req:     requests.TagRequest{Name: "Gin"},
			saved:   models.Tag{Id: "t1", Name: "Gin", UserId: "u1", Approved: true},
			wantTag: models.Tag{Id: "t1", Name: "Gin", UserId: "u1", Approved: true},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, saved models.Tag) {
				repo.EXPECT().
					CreateTag(ctx, mock.MatchedBy(func(tag models.Tag) bool {
						return tag.Name == "Gin" && tag.UserId == "u1" && tag.Approved
					})).
					Return(saved, nil)
			},
		},
		{
			name:    "regular user -> tag not approved",
			user:    models.User{Id: "u2", Role: 0},
			req:     requests.TagRequest{Name: "Rum"},
			saved:   models.Tag{Id: "t2", Name: "Rum", UserId: "u2", Approved: false},
			wantTag: models.Tag{Id: "t2", Name: "Rum", UserId: "u2", Approved: false},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, saved models.Tag) {
				repo.EXPECT().
					CreateTag(ctx, mock.MatchedBy(func(tag models.Tag) bool {
						return tag.Name == "Rum" && tag.UserId == "u2" && !tag.Approved
					})).
					Return(saved, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc // захват переменной
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)

			ctx := models.WithUser(context.Background(), tc.user)
			tc.setupMock(repo, ctx, tc.saved)

			got, err := svc.CreateTag(ctx, tc.req)

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTag, got)
		})
	}
}

func TestTagService_AddTagToCocktail(t *testing.T) {
	tests := []struct {
		name      string
		setupMock func(repo *mocks.TagRepository, ctx context.Context, usr models.User)
		wantErr   bool
		wantTag   models.Tag
		user      models.User
	}{
		{
			name: "тег уже существует и утверждён",
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, _ models.User) {
				tag := models.Tag{Id: "t_exist", Name: "Sweet", Approved: true}
				repo.EXPECT().
					IsTagExistAndApproved(ctx, "Sweet").
					Return(true, tag, nil)
				repo.EXPECT().
					InsertCocktailTag(ctx, "cocktail-1", "t_exist").
					Return(nil)
			},
			wantErr: false,
			wantTag: models.Tag{Id: "t_exist", Name: "Sweet", Approved: true},
			user:    models.User{Id: "u1", Role: 0},
		},
		{
			name: "тега нет → создаём новый",
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, usr models.User) {
				newTag := models.Tag{Id: "t_new", Name: "Bitter", UserId: usr.Id, Approved: false}

				repo.EXPECT().
					IsTagExistAndApproved(ctx, "Bitter").
					Return(false, models.Tag{}, nil)

				// ожидание вызова repo.CreateTag изнутри svc.CreateTag(...)
				repo.EXPECT().
					CreateTag(ctx, mock.MatchedBy(func(tag models.Tag) bool {
						return tag.Name == "Bitter" && tag.UserId == usr.Id && !tag.Approved
					})).
					Return(newTag, nil)

				repo.EXPECT().
					InsertCocktailTag(ctx, "cocktail-1", "t_new").
					Return(nil)
			},
			wantErr: false,
			wantTag: models.Tag{Id: "t_new", Name: "Bitter", UserId: "u1", Approved: false},
			user:    models.User{Id: "u1", Role: 0},
		},
		{
			name: "IsTagExistAndApproved вернул ошибку",
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, _ models.User) {
				repo.EXPECT().
					IsTagExistAndApproved(ctx, "Dry").
					Return(false, models.Tag{}, errors.New("db fail"))
			},
			wantTag: models.Tag{Id: "t_exist", Name: "Dry", UserId: "u1", Approved: true},
			wantErr: true,
			user:    models.User{Id: "u1", Role: 0},
		},
		{
			name: "InsertCocktailTag вернул ошибку",
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, _ models.User) {
				tag := models.Tag{Id: "t_exist2", Name: "Fresh", Approved: true}

				repo.EXPECT().
					IsTagExistAndApproved(ctx, "Fresh").
					Return(true, tag, nil)

				repo.EXPECT().
					InsertCocktailTag(ctx, "cocktail-1", "t_exist2").
					Return(errors.New("duplicate row"))
			},
			wantTag: models.Tag{Id: "t_exist2", Name: "Fresh"},
			wantErr: true,
			user:    models.User{Id: "u1", Role: 0},
		},
	}

	for _, tc := range tests {
		tc := tc // захват переменной
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			ctx := models.WithUser(context.Background(), tc.user)

			// индивидуальная подготовка мока
			tc.setupMock(repo, ctx, tc.user)

			got, err := svc.AddTagToCocktail(ctx, "cocktail-1",
				requests.TagRequest{Name: tc.wantTag.Name})

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTag, got)
		})
	}
}

func TestTagService_GetTagsByCocktailID(t *testing.T) {
	tests := []struct {
		name      string
		saved     []models.Tag // то, что «репо» вернёт
		wantErr   bool
		wantTag   []models.Tag // ожидаемый результат сервиса
		setupMock func(*mocks.TagRepository, context.Context, string, []models.Tag)
	}{
		{
			name: "теги есть у данного коктейля",
			saved: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u2", Approved: true},
			},
			wantErr: false,
			wantTag: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u2", Approved: true},
			},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, cocktailId string, saved []models.Tag) {
				repo.EXPECT().
					GetTagsByCocktailID(ctx, cocktailId).
					Return(saved, nil)
			},
		},
		{
			name:    "тегов нет у данного коктейля",
			saved:   []models.Tag{},
			wantErr: false,
			wantTag: []models.Tag{},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, cocktailId string, saved []models.Tag) {
				repo.EXPECT().
					GetTagsByCocktailID(ctx, cocktailId).
					Return(saved, nil)
			},
		},
		{
			name: "GetTagsByCocktailID вернул ошибку",
			saved: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u2", Approved: true},
			},
			wantErr: true,
			wantTag: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u2", Approved: true},
			},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, cocktailId string, saved []models.Tag) {
				repo.EXPECT().
					GetTagsByCocktailID(ctx, cocktailId).
					Return([]models.Tag{}, errors.New("db fail"))
			},
		},
	}

	for _, tc := range tests {
		tc := tc // захват переменной
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			ctx := context.Background()

			// индивидуальная подготовка мока
			tc.setupMock(repo, ctx, "cocktail-1", tc.saved)

			got, err := svc.GetTagsByCocktailID(ctx, "cocktail-1")

			if tc.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTag, got)
		})
	}
}

func TestTagService_GetPersonalTags(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		saved     []models.Tag // то, что «репо» вернёт
		wantErr   string
		wantTag   []models.Tag // ожидаемый результат сервиса
		setupMock func(*mocks.TagRepository, context.Context, string, []models.Tag)
	}{
		{
			name: "получение персональных тегов",
			ctx:  models.WithUser(context.Background(), models.User{Id: "u1"}),
			saved: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: true},
			},
			wantErr: "",
			wantTag: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: true},
			},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, userId string, saved []models.Tag) {
				repo.EXPECT().
					GetPersonalTags(ctx, userId).
					Return(saved, nil)
			},
		},
		{
			name:    "тегов нет у данного коктейля",
			ctx:     models.WithUser(context.Background(), models.User{Id: "u1"}),
			saved:   []models.Tag{},
			wantErr: "",
			wantTag: []models.Tag{},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, userId string, saved []models.Tag) {
				repo.EXPECT().
					GetPersonalTags(ctx, userId).
					Return(saved, nil)
			},
		},
		{
			name:      "пользователь не авторизован",
			ctx:       context.Background(),
			saved:     []models.Tag{},
			wantErr:   "unathorized",
			wantTag:   []models.Tag{},
			setupMock: nil,
		},
		{
			name: "GetPersonalTags вернул ошибку",
			ctx:  models.WithUser(context.Background(), models.User{Id: "u1"}),
			saved: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: true},
			},
			wantErr: "getpersonaltags",
			wantTag: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: true},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: true},
			},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, userId string, saved []models.Tag) {
				repo.EXPECT().
					GetPersonalTags(ctx, userId).
					Return([]models.Tag{}, errors.New("db fail"))
			},
		},
	}

	for _, tc := range tests {
		tc := tc // захват переменной
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			user, _ := models.GetUserFromContext(tc.ctx)
			// индивидуальная подготовка мока
			if tc.setupMock != nil {
				tc.setupMock(repo, tc.ctx, user.Id, tc.saved)
			}

			got, err := svc.GetPersonalTags(tc.ctx)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Len(t, got, 0)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTag, got)
		})
	}
}

func TestTagService_GetUnapprovedTags(t *testing.T) {
	tests := []struct {
		name      string
		ctx       context.Context
		saved     []models.Tag // то, что «репо» вернёт
		wantErr   string
		wantTag   []models.Tag // ожидаемый результат сервиса
		setupMock func(*mocks.TagRepository, context.Context, []models.Tag)
	}{
		{
			name: "получение неапрувнутых тегов",
			ctx:  models.WithUser(context.Background(), models.User{Id: "u1", Role: 1}),
			saved: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: false},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: false},
			},
			wantErr: "",
			wantTag: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: false},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: false},
			},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, saved []models.Tag) {
				repo.EXPECT().
					GetUnapprovedTags(ctx).
					Return(saved, nil)
			},
		},
		{
			name:    "тегов нет",
			ctx:     models.WithUser(context.Background(), models.User{Id: "u1", Role: 1}),
			saved:   []models.Tag{},
			wantErr: "",
			wantTag: []models.Tag{},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, saved []models.Tag) {
				repo.EXPECT().
					GetUnapprovedTags(ctx).
					Return(saved, nil)
			},
		},
		{
			name:      "пользователь не админ",
			ctx:       models.WithUser(context.Background(), models.User{Id: "u1", Role: 0}),
			saved:     []models.Tag{},
			wantErr:   "restricted access",
			wantTag:   []models.Tag{},
			setupMock: nil,
		},
		{
			name:      "пользователь не авторизован",
			ctx:       context.Background(),
			saved:     []models.Tag{},
			wantErr:   "unathorized",
			wantTag:   []models.Tag{},
			setupMock: nil,
		},
		{
			name: "GetUnapprovedTags вернул ошибку",
			ctx:  models.WithUser(context.Background(), models.User{Id: "u1", Role: 1}),
			saved: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: false},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: false},
			},
			wantErr: "getunapprovedtags",
			wantTag: []models.Tag{
				{Id: "t_exist1", Name: "Dry", UserId: "u1", Approved: false},
				{Id: "t_exist2", Name: "Sour", UserId: "u1", Approved: false},
			},
			setupMock: func(repo *mocks.TagRepository, ctx context.Context, saved []models.Tag) {
				repo.EXPECT().
					GetUnapprovedTags(ctx).
					Return([]models.Tag{}, errors.New("db fail"))
			},
		},
	}

	for _, tc := range tests {
		tc := tc // захват переменной
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			// индивидуальная подготовка мока
			if tc.setupMock != nil {
				tc.setupMock(repo, tc.ctx, tc.saved)
			}

			got, err := svc.GetUnapprovedTags(tc.ctx)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Len(t, got, 0)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantTag, got)
		})
	}
}

func TestTagService_UpdateTagsApprovedStatus(t *testing.T) {
	updateReq := []requests.TagUpdateRequest{
		{Id: "t1", Approved: true},
		{Id: "t2", Approved: false},
	}
	expectedMapped := []models.Tag{
		{Id: "t1", Approved: true},
		{Id: "t2", Approved: false},
	}

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(r *mocks.TagRepository, ctx context.Context)
		wantErr string // "" => ошибки не ждём
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name:    "not admin",
			ctx:     models.WithUser(context.Background(), models.User{Id: "u1", Role: 0}),
			wantErr: "restricted access",
		},
		{
			name: "admin OK",
			ctx:  models.WithUser(context.Background(), models.User{Id: "admin", Role: 1}),
			setup: func(r *mocks.TagRepository, ctx context.Context) {
				// проверяем, что в репозиторий ушёл именно ожидаемый срез
				r.EXPECT().
					UpdateTagsApprovedStatus(ctx,
						mock.MatchedBy(func(tags []models.Tag) bool {
							return assert.ObjectsAreEqual(expectedMapped, tags)
						}),
					).Return(nil)
			},
		},
		{
			name: "admin repo error",
			ctx:  models.WithUser(context.Background(), models.User{Id: "admin", Role: 1}),
			setup: func(r *mocks.TagRepository, ctx context.Context) {
				r.EXPECT().
					UpdateTagsApprovedStatus(ctx, mock.Anything).
					Return(errors.New("db fail"))
			},
			wantErr: "failed to update tags",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			err := svc.UpdateTagsApprovedStatus(tc.ctx, updateReq)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTagService_DeleteTag(t *testing.T) {
	const tagID = "tag-1"

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(r *mocks.TagRepository, ctx context.Context) // nil ⇒ ожиданий нет
		wantErr string                                            // "" ⇒ ошибки не ждём
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name:    "not admin",
			ctx:     models.WithUser(context.Background(), models.User{Id: "user", Role: 0}),
			wantErr: "restricted access",
		},
		{
			name: "admin OK",
			ctx:  models.WithUser(context.Background(), models.User{Id: "admin", Role: 1}),
			setup: func(r *mocks.TagRepository, ctx context.Context) {
				r.EXPECT().
					DeleteTag(ctx, tagID).
					Return(nil)
			},
		},
		{
			name: "admin repo error",
			ctx:  models.WithUser(context.Background(), models.User{Id: "admin", Role: 1}),
			setup: func(r *mocks.TagRepository, ctx context.Context) {
				r.EXPECT().
					DeleteTag(ctx, tagID).
					Return(errors.New("db fail"))
			},
			wantErr: "failed to delete tag",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			err := svc.DeleteTag(tc.ctx, tagID)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestTagService_DeleteTagFromCocktail(t *testing.T) {
	const (
		tagID      = "tag-1"
		cocktailID = "cocktail-1"
	)

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(r *mocks.TagRepository, ctx context.Context)
		wantErr string // "" → ошибки не ждём; иначе подстроку ищем в err
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(), // без пользователя
			wantErr: "unathorized",
		},
		{
			name: "success",
			ctx:  models.WithUser(context.Background(), models.User{Id: "u1", Role: 0}),
			setup: func(r *mocks.TagRepository, ctx context.Context) {
				r.EXPECT().
					DeleteCocktailTag(ctx, cocktailID, tagID, "u1", 0).
					Return(nil)
			},
		},
		{
			name: "repository error",
			ctx:  models.WithUser(context.Background(), models.User{Id: "admin", Role: 1}),
			setup: func(r *mocks.TagRepository, ctx context.Context) {
				r.EXPECT().
					DeleteCocktailTag(ctx, cocktailID, tagID, "admin", 1).
					Return(errors.New("db fail"))
			},
			wantErr: "failed to delete tag",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newTagService(t)
			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			err := svc.DeleteTagFromCocktail(tc.ctx, tagID, cocktailID)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}
