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
	"github.com/stretchr/testify/require"
)

func newCocktailSvc(t *testing.T) (*CocktailService, *mocks.CocktailRepository, *mocks.CsIngredientRepository, *mocks.CsTagRepository, *mocks.CsIngredientNomRepository) {
	repoCocktail := mocks.NewCocktailRepository(t)
	repoIngr := mocks.NewCsIngredientRepository(t)
	repoTag := mocks.NewCsTagRepository(t)
	repoIngrNom := mocks.NewCsIngredientNomRepository(t)
	return NewCocktailService(repoCocktail, repoIngr, repoTag, repoIngrNom), repoCocktail, repoIngr, repoTag, repoIngrNom
}

func TestCreateCocktail(t *testing.T) {
	adminCtx := ctxWithUser("admin", 1)
	userCtx := ctxWithUser("user1", 0)

	minReq := requests.CocktailRequest{
		Name:    "Negroni",
		Picture: "pic.png",
		Rating:  4.5,
		// Ingredients, Tags — пустые, чтобы не тащить длинные мок-цепочки
	}

	tests := []struct {
		name  string
		ctx   context.Context
		req   requests.CocktailRequest
		setup func(cockRepo *mocks.CocktailRepository,
			nomRepo *mocks.CsIngredientNomRepository,
			tagRepo *mocks.CsTagRepository)
		wantErr string
	}{
		// ───────────────────── unauthorized ─────────────────────
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			req:     minReq,
			wantErr: "unathorized",
		},
		// ───────────────────── duplicate cocktail ────────────────
		{
			name: "already exists",
			ctx:  userCtx,
			req:  minReq,
			setup: func(cockRepo *mocks.CocktailRepository,
				_ *mocks.CsIngredientNomRepository,
				_ *mocks.CsTagRepository) {

				cockRepo.EXPECT().
					IsCocktailExistAndApproved(userCtx, "Negroni").
					Return(true, nil) // уже есть
			},
			wantErr: "cocktail has been already created",
		},
		// ───────────────────── repo error in existence check ─────
		{
			name: "IsCocktailExistAndApproved error",
			ctx:  userCtx,
			req:  minReq,
			setup: func(cockRepo *mocks.CocktailRepository,
				_ *mocks.CsIngredientNomRepository,
				_ *mocks.CsTagRepository) {

				cockRepo.EXPECT().
					IsCocktailExistAndApproved(userCtx, "Negroni").
					Return(false, errors.New("db fail"))
			},
			wantErr: "iscocktailexist",
		},
		// ───────────────────── success (admin, no ingredients/tags) ─────
		{
			name: "admin creates cocktail (approved immediately)",
			ctx:  adminCtx,
			req:  minReq,
			setup: func(cockRepo *mocks.CocktailRepository,
				_ *mocks.CsIngredientNomRepository,
				_ *mocks.CsTagRepository) {

				cockRepo.EXPECT().
					IsCocktailExistAndApproved(adminCtx, "Negroni").
					Return(false, nil)

				cockRepo.EXPECT().
					CreateCocktail(adminCtx,
						mock.MatchedBy(func(c models.Cocktail) bool {
							return c.Name == "Negroni" && c.Approved
						}),
					).
					Return(models.Cocktail{Id: "c1", Name: "Negroni", Approved: true}, nil)
			},
		},
		// ───────────────────── success (regular user, pending) ─────
		{
			name: "regular user creates cocktail (pending)",
			ctx:  userCtx,
			req:  minReq,
			setup: func(cockRepo *mocks.CocktailRepository,
				_ *mocks.CsIngredientNomRepository,
				_ *mocks.CsTagRepository) {

				cockRepo.EXPECT().
					IsCocktailExistAndApproved(userCtx, "Negroni").
					Return(false, nil)

				cockRepo.EXPECT().
					CreateCocktail(userCtx,
						mock.MatchedBy(func(c models.Cocktail) bool {
							return c.Name == "Negroni" && !c.Approved
						}),
					).
					Return(models.Cocktail{Id: "c2", Name: "Negroni", Approved: false}, nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, cockRepo, _, tagRepo, nomRepo := newCocktailSvc(t)

			if tc.setup != nil {
				tc.setup(cockRepo, nomRepo, tagRepo)
			}

			got, err := svc.CreateCocktail(tc.ctx, tc.req)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "Negroni", got.Name)
		})
	}
}

func TestUpdateCocktail(t *testing.T) {
	adminCtx := ctxWithUser("admin", 1)
	ownerCtx := ctxWithUser("user1", 0)
	otherCtx := ctxWithUser("user2", 0)

	req := requests.CocktailRequest{
		Name:   "Updated Negroni",
		Rating: 4.8,
	}

	tests := []struct {
		name     string
		ctx      context.Context
		userId   string // нужен для мокировки user.Id в mapCocktail
		role     int
		setup    func(repo *mocks.CocktailRepository, expected models.Cocktail)
		wantErr  string
		mockUser string
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name:     "admin updates successfully",
			ctx:      adminCtx,
			mockUser: "admin",
			setup: func(repo *mocks.CocktailRepository, cocktail models.Cocktail) {
				repo.EXPECT().UpdateCocktail(adminCtx, cocktail).Return(nil)
			},
		},
		{
			name:     "owner updates successfully",
			ctx:      ownerCtx,
			mockUser: "user1",
			setup: func(repo *mocks.CocktailRepository, cocktail models.Cocktail) {
				repo.EXPECT().UpdateCocktail(ownerCtx, cocktail).Return(nil)
			},
		},
		{
			name:     "other user — restricted",
			ctx:      otherCtx,
			mockUser: "user1",
			wantErr:  "restricted access",
		},
		{
			name:     "repo error on update",
			ctx:      adminCtx,
			mockUser: "admin",
			setup: func(repo *mocks.CocktailRepository, cocktail models.Cocktail) {
				repo.EXPECT().UpdateCocktail(adminCtx, cocktail).Return(errors.New("db error"))
			},
			wantErr: "db error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, cockRepo, _, _, _ := newCocktailSvc(t)

			// --- Переопределяем mapCocktail ---
			original := mapCocktail
			mapCocktail = func(_ string, req requests.CocktailRequest) models.Cocktail {
				return models.Cocktail{
					Name:   req.Name,
					Rating: req.Rating,
					UserId: tt.mockUser, // то, что мы задали в таблице
				}
			}
			defer func() { mapCocktail = original }()

			// задаём ожидания, если нужны
			if tt.setup != nil {
				expected := mapCocktail("", req)
				tt.setup(cockRepo, expected)
			}

			err := svc.UpdateCocktail(tt.ctx, req)

			if tt.wantErr != "" {
				assert.ErrorContains(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGetCocktailsByName(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name           string
		query          string
		setupMocks     func(*mocks.CocktailRepository, *mocks.CsIngredientRepository, *mocks.CsTagRepository)
		expectedResult []models.Cocktail
		expectedError  string
	}{
		{
			name:  "success - 1 cocktail with ingredients and tags",
			query: "Negroni",
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, tagRepo *mocks.CsTagRepository) {
				cocktails := []models.Cocktail{
					{Id: "c1", Name: "Negroni"},
				}

				cockRepo.EXPECT().GetCocktailsByName(ctx, "Negroni").Return(cocktails, nil)

				ingrRepo.EXPECT().GetIngredientsByCocktailID(ctx, "c1").Return([]models.Ingredient{
					{Name: models.IngredientNomenclature{Name: "Gin"}},
				}, nil)

				tagRepo.EXPECT().GetTagsByCocktailID(ctx, "c1").Return([]models.Tag{
					{Name: "Classic"},
				}, nil)
			},
			expectedResult: []models.Cocktail{
				{
					Id:   "c1",
					Name: "Negroni",
					Ingredients: []models.Ingredient{
						{Name: models.IngredientNomenclature{Name: "Gin"}},
					},
					Tags: []models.Tag{
						{Name: "Classic"},
					},
				},
			},
		},
		{
			name:  "cocktail repo error",
			query: "Mojito",
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, tagRepo *mocks.CsTagRepository) {
				cockRepo.EXPECT().GetCocktailsByName(ctx, "Mojito").Return(nil, errors.New("db error"))
			},
			expectedError: "db error",
		},
		{
			name:  "ingredient repo error",
			query: "Old Fashioned",
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, tagRepo *mocks.CsTagRepository) {
				cockRepo.EXPECT().GetCocktailsByName(ctx, "Old Fashioned").Return([]models.Cocktail{
					{Id: "c2", Name: "Old Fashioned"},
				}, nil)

				ingrRepo.EXPECT().GetIngredientsByCocktailID(ctx, "c2").Return(nil, errors.New("ingredient error"))
			},
			expectedError: "ingredient error",
		},
		{
			name:  "tag repo error",
			query: "Martini",
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, tagRepo *mocks.CsTagRepository) {
				cockRepo.EXPECT().GetCocktailsByName(ctx, "Martini").Return([]models.Cocktail{
					{Id: "c3", Name: "Martini"},
				}, nil)

				ingrRepo.EXPECT().GetIngredientsByCocktailID(ctx, "c3").Return([]models.Ingredient{}, nil)
				tagRepo.EXPECT().GetTagsByCocktailID(ctx, "c3").Return(nil, errors.New("tag error"))
			},
			expectedError: "tag error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, cockRepo, ingrRepo, tagRepo, _ := newCocktailSvc(t)

			if tt.setupMocks != nil {
				tt.setupMocks(cockRepo, ingrRepo, tagRepo)
			}

			result, err := svc.GetCocktailsByName(ctx, tt.query)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedResult, result)
			}
		})
	}
}

func TestGetCocktails(t *testing.T) {
	ctx := context.Background()
	user := models.User{Id: "u123", Role: 0}
	ctxWithUser := models.WithUser(ctx, user)

	tests := []struct {
		name          string
		ctx           context.Context
		isApproved    bool
		setupMocks    func(*mocks.CocktailRepository, *mocks.CsIngredientRepository, *mocks.CsTagRepository)
		expected      []models.Cocktail
		expectedError string
	}{
		{
			name:          "unauthorized user",
			ctx:           ctx,
			isApproved:    true,
			expected:      []models.Cocktail{},
			expectedError: "unathorized",
		},
		{
			name:       "success - single cocktail with tags and ingredients",
			ctx:        ctxWithUser,
			isApproved: true,
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, tagRepo *mocks.CsTagRepository) {
				cocktails := []models.Cocktail{
					{Id: "c1", Name: "Negroni"},
				}
				cockRepo.EXPECT().GetCocktails(ctxWithUser, true, user.Id).Return(cocktails, nil)

				ingrRepo.EXPECT().GetIngredientsByCocktailID(ctxWithUser, "c1").Return([]models.Ingredient{
					{Name: models.IngredientNomenclature{Name: "Gin"}},
				}, nil)

				tagRepo.EXPECT().GetTagsByCocktailID(ctxWithUser, "c1").Return([]models.Tag{
					{Name: "Bitter"},
				}, nil)
			},
			expected: []models.Cocktail{
				{
					Id:   "c1",
					Name: "Negroni",
					Ingredients: []models.Ingredient{
						{Name: models.IngredientNomenclature{Name: "Gin"}},
					},
					Tags: []models.Tag{
						{Name: "Bitter"},
					},
				},
			},
		},
		{
			name:       "error in GetCocktails",
			ctx:        ctxWithUser,
			isApproved: false,
			setupMocks: func(cockRepo *mocks.CocktailRepository, _ *mocks.CsIngredientRepository, _ *mocks.CsTagRepository) {
				cockRepo.EXPECT().GetCocktails(ctxWithUser, false, user.Id).Return(nil, errors.New("db error"))
			},
			expectedError: "db error",
		},
		{
			name:       "error in GetIngredientsByCocktailID",
			ctx:        ctxWithUser,
			isApproved: true,
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, _ *mocks.CsTagRepository) {
				cockRepo.EXPECT().GetCocktails(ctxWithUser, true, user.Id).Return([]models.Cocktail{
					{Id: "c2", Name: "Martini"},
				}, nil)

				ingrRepo.EXPECT().GetIngredientsByCocktailID(ctxWithUser, "c2").Return(nil, errors.New("ingredient error"))
			},
			expectedError: "ingredient error",
		},
		{
			name:       "error in GetTagsByCocktailID",
			ctx:        ctxWithUser,
			isApproved: true,
			setupMocks: func(cockRepo *mocks.CocktailRepository, ingrRepo *mocks.CsIngredientRepository, tagRepo *mocks.CsTagRepository) {
				cockRepo.EXPECT().GetCocktails(ctxWithUser, true, user.Id).Return([]models.Cocktail{
					{Id: "c3", Name: "Old Fashioned"},
				}, nil)

				ingrRepo.EXPECT().GetIngredientsByCocktailID(ctxWithUser, "c3").Return([]models.Ingredient{}, nil)
				tagRepo.EXPECT().GetTagsByCocktailID(ctxWithUser, "c3").Return(nil, errors.New("tag error"))
			},
			expectedError: "tag error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, cockRepo, ingrRepo, tagRepo, _ := newCocktailSvc(t)

			if tt.setupMocks != nil {
				tt.setupMocks(cockRepo, ingrRepo, tagRepo)
			}

			got, err := svc.GetCocktails(tt.ctx, tt.isApproved)

			if tt.expectedError != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.expectedError)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expected, got)
			}
		})
	}
}

func TestSearchCocktails(t *testing.T) {
	tests := []struct {
		name               string
		ingredients        []requests.IngredientRequest
		tags               []requests.TagRequest
		pagination         models.Pagination
		mockIngrResult     []models.Cocktail
		mockTagResult      []models.Cocktail
		expectedCocktailID []string
	}{
		{
			name:        "intersection success with pagination",
			ingredients: []requests.IngredientRequest{{Name: "Mint"}},
			tags:        []requests.TagRequest{{Name: "Fresh"}},
			pagination:  models.Pagination{PageNumber: 1, PageSize: 1},
			mockIngrResult: []models.Cocktail{
				{Id: "c1", Name: "Mojito"},
				{Id: "c2", Name: "Margarita"},
			},
			mockTagResult: []models.Cocktail{
				{Id: "c2", Name: "Margarita"},
			},
			expectedCocktailID: []string{"c2"},
		},
		{
			name:        "no intersection returns empty result",
			ingredients: []requests.IngredientRequest{{Name: "Mint"}},
			tags:        []requests.TagRequest{{Name: "Bitter"}},
			pagination:  models.Pagination{PageNumber: 1, PageSize: 10},
			mockIngrResult: []models.Cocktail{
				{Id: "c1", Name: "Mojito"},
			},
			mockTagResult: []models.Cocktail{
				{Id: "c2", Name: "Negroni"},
			},
			expectedCocktailID: []string{},
		},
		{
			name:        "pagination out of range",
			ingredients: []requests.IngredientRequest{{Name: "Mint"}},
			tags:        []requests.TagRequest{{Name: "Fresh"}},
			pagination:  models.Pagination{PageNumber: 10, PageSize: 5},
			mockIngrResult: []models.Cocktail{
				{Id: "c1"},
			},
			mockTagResult: []models.Cocktail{
				{Id: "c1"},
			},
			expectedCocktailID: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, cockRepo, _, _, _ := newCocktailSvc(t)
			cockRepo.EXPECT().GetCocktailsByIngredients(mock.Anything, mock.Anything).Return(tt.mockIngrResult, nil)
			cockRepo.EXPECT().GetCocktailsByTags(mock.Anything, mock.Anything).Return(tt.mockTagResult, nil)

			got, err := svc.SearchCocktails(context.Background(), "", tt.ingredients, tt.tags, tt.pagination)

			assert.NoError(t, err)
			assert.Len(t, got, len(tt.expectedCocktailID))
			for i, c := range got {
				assert.Equal(t, tt.expectedCocktailID[i], c.Id)
			}
		})
	}
}

func TestDeleteCocktail(t *testing.T) {
	type testCase struct {
		name          string
		ctx           context.Context
		user          models.User
		cocktail      models.Cocktail
		isAuthorized  bool
		isAdmin       bool
		expectedError string
		mockError     error
	}

	tests := []testCase{
		{
			name:         "success by admin",
			user:         models.User{Id: "admin", Role: 1},
			cocktail:     models.Cocktail{Id: "c1", UserId: "u123"},
			isAuthorized: true,
			isAdmin:      true,
		},
		{
			name:         "success by owner",
			user:         models.User{Id: "u123", Role: 0},
			cocktail:     models.Cocktail{Id: "c1", UserId: "u123"},
			isAuthorized: true,
		},
		{
			name:          "unauthorized user",
			isAuthorized:  false,
			expectedError: "unathorized",
		},
		{
			name:          "access restricted for non-owner",
			user:          models.User{Id: "u123", Role: 0},
			cocktail:      models.Cocktail{Id: "c1", UserId: "someone_else"},
			isAuthorized:  true,
			expectedError: "access restricted",
		},
		{
			name:          "repo error on GetCocktailByID",
			user:          models.User{Id: "admin", Role: 1},
			isAuthorized:  true,
			mockError:     errors.New("repo fail"),
			expectedError: "repo fail",
		},
		{
			name:          "repo error on DeleteCocktail",
			user:          models.User{Id: "admin", Role: 1},
			cocktail:      models.Cocktail{Id: "c1", UserId: "u123"},
			isAuthorized:  true,
			mockError:     errors.New("delete error"),
			expectedError: "delete error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc, cockRepo, _, _, _ := newCocktailSvc(t)

			ctx := context.Background()
			if tt.isAuthorized {
				ctx = models.WithUser(ctx, tt.user)
			}

			if tt.expectedError != "unathorized" {
				cockRepo.EXPECT().
					GetCocktailByID(ctx, "c1").
					Return(tt.cocktail, tt.mockError).
					Maybe()

				if tt.mockError == nil && (tt.isAdmin || tt.user.Id == tt.cocktail.UserId) {
					cockRepo.EXPECT().
						DeleteCocktail(ctx, "c1").
						Return(tt.mockError).
						Maybe()
				}
			}

			err := svc.DeleteCocktail(ctx, "c1")

			if tt.expectedError != "" {
				require.EqualError(t, err, tt.expectedError)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
