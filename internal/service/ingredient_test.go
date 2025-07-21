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

func newIngredientSvc(t *testing.T) (*IngredientService, *mocks.IngredientRepository, *mocks.IngrtNomRepository) {
	repoIngr := mocks.NewIngredientRepository(t)
	repoIngrNom := mocks.NewIngrtNomRepository(t)
	return NewIngredientService(repoIngr, repoIngrNom), repoIngr, repoIngrNom
}

func ctxWithUser(id string, role int) context.Context {
	u := models.User{Id: id, Role: role}
	return models.WithUser(context.Background(), u)
}

func TestCreateIngredient(t *testing.T) {
	req := requests.IngredientRequest{
		Name:    "Gin",
		Amount:  30,
		Measure: "ml",
		// Replacement  nil ‒ по умолчанию
	}

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(r *mocks.IngrtNomRepository, ctx context.Context)
		wantErr string
		check   func(t *testing.T, ingr models.Ingredient)
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name: "nom exists & approved",
			ctx:  ctxWithUser("u1", 0),
			setup: func(r *mocks.IngrtNomRepository, ctx context.Context) {
				nom := models.IngredientNomenclature{Id: "nom1", Name: "Gin", Approved: true}
				r.EXPECT().
					IsIngredientNomExistAndApproved(ctx, "Gin").
					Return(true, nom, nil)
			},
			check: func(t *testing.T, ingr models.Ingredient) {
				assert.Equal(t, "nom1", ingr.Name.Id)
				assert.Equal(t, float32(30), ingr.Amount)
				assert.Equal(t, "ml", ingr.Measure)
			},
		},
		{
			name: "nom not exist, admin creates (approved)",
			ctx:  ctxWithUser("admin", 1),
			setup: func(r *mocks.IngrtNomRepository, ctx context.Context) {
				r.EXPECT().
					IsIngredientNomExistAndApproved(ctx, "Gin").
					Return(false, models.IngredientNomenclature{}, nil)

				r.EXPECT().
					CreateIngredientNom(ctx,
						mock.MatchedBy(func(n models.IngredientNomenclature) bool {
							return n.Name == "Gin" && n.Approved
						}),
					).
					Return(models.IngredientNomenclature{Id: "newNom", Name: "Gin", Approved: true}, nil)
			},
			check: func(t *testing.T, ingr models.Ingredient) {
				assert.Equal(t, "newNom", ingr.Name.Id)
				assert.True(t, ingr.Name.Approved)
			},
		},
		{
			name: "repo error on IsIngredientNomExistAndApproved",
			ctx:  ctxWithUser("u1", 0),
			setup: func(r *mocks.IngrtNomRepository, ctx context.Context) {
				r.EXPECT().
					IsIngredientNomExistAndApproved(ctx, "Gin").
					Return(false, models.IngredientNomenclature{}, errors.New("db fail"))
			},
			wantErr: "isingredientnomexist",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, _, repoNom := newIngredientSvc(t)

			if tc.setup != nil {
				tc.setup(repoNom, tc.ctx)
			}

			got, err := svc.CreateIngredient(tc.ctx, "cocktail-1", req)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			if tc.check != nil {
				tc.check(t, got)
			}
		})
	}
}

func TestAddIngredientToCocktail(t *testing.T) {
	const cocktailID = "cocktail-1"

	req := requests.IngredientRequest{
		Name:    "Gin",
		Amount:  30,
		Measure: "ml",
	}

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(ingrRepo *mocks.IngredientRepository, nomRepo *mocks.IngrtNomRepository, ctx context.Context)
		wantErr string
	}{
		// ───────── create-ingredient error ─────────
		{
			name: "CreateIngredient returns error",
			ctx:  ctxWithUser("u1", 0),
			setup: func(_ *mocks.IngredientRepository, nomRepo *mocks.IngrtNomRepository, ctx context.Context) {
				// IsIngredientNomExistAndApproved вернёт ошибку -> CreateIngredient упадёт
				nomRepo.EXPECT().
					IsIngredientNomExistAndApproved(ctx, "Gin").
					Return(false, models.IngredientNomenclature{}, errors.New("db fail"))
				// InsertIngredient НЕ должен вызываться (AssertExpectations проверит)
			},
			wantErr: "create ingredient",
		},
		// ───────── insert error ─────────
		{
			name: "InsertIngredient returns error",
			ctx:  ctxWithUser("admin", 1), // админ создаёт и ном, и ингредиент
			setup: func(ingrRepo *mocks.IngredientRepository, nomRepo *mocks.IngrtNomRepository, ctx context.Context) {
				// 1) CreateIngredient внутри:
				nomRepo.EXPECT().
					IsIngredientNomExistAndApproved(ctx, "Gin").
					Return(false, models.IngredientNomenclature{}, nil)
				nomRepo.EXPECT().
					CreateIngredientNom(ctx,
						mock.MatchedBy(func(n models.IngredientNomenclature) bool {
							return n.Name == "Gin"
						})).
					Return(models.IngredientNomenclature{Id: "nom1", Name: "Gin"}, nil)

				// 2) InsertIngredient отдаёт ошибку
				ingrRepo.EXPECT().
					InsertIngredient(ctx,
						mock.MatchedBy(func(ingr models.Ingredient) bool { return ingr.Name.Name == "Gin" }),
						cocktailID).
					Return(errors.New("duplicate"))
			},
			wantErr: "insert ingredient",
		},
		// ───────── happy-path ─────────
		{
			name: "success",
			ctx:  ctxWithUser("u1", 0),
			setup: func(ingrRepo *mocks.IngredientRepository, nomRepo *mocks.IngrtNomRepository, ctx context.Context) {
				nom := models.IngredientNomenclature{Id: "nom2", Name: "Gin", Approved: true}

				nomRepo.EXPECT().
					IsIngredientNomExistAndApproved(ctx, "Gin").
					Return(true, nom, nil)

				ingrRepo.EXPECT().
					InsertIngredient(ctx,
						mock.MatchedBy(func(ingr models.Ingredient) bool {
							return ingr.Name.Id == "nom2" && ingr.Amount == 30
						}),
						cocktailID).
					Return(nil)
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, ingrRepo, nomRepo := newIngredientSvc(t)
			if tc.setup != nil {
				tc.setup(ingrRepo, nomRepo, tc.ctx)
			}

			got, err := svc.AddIngredientToCocktail(tc.ctx, cocktailID, req)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, "Gin", got.Name.Name)
			assert.Equal(t, float32(30), got.Amount)
		})
	}
}
func TestDeleteIngredientFromCocktail(t *testing.T) {
	const (
		cocktailID   = "cocktail-1"
		ingredientID = "ingr-1"
	)

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *mocks.IngredientRepository, ctx context.Context)
		wantErr string
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name: "success",
			ctx:  ctxWithUser("u1", 0),
			setup: func(repo *mocks.IngredientRepository, ctx context.Context) {
				repo.EXPECT().
					DeleteIngredient(ctx, ingredientID, cocktailID, "u1").
					Return(nil)
			},
		},
		{
			name: "repo error",
			ctx:  ctxWithUser("admin", 1),
			setup: func(repo *mocks.IngredientRepository, ctx context.Context) {
				repo.EXPECT().
					DeleteIngredient(ctx, ingredientID, cocktailID, "admin").
					Return(errors.New("db fail"))
			},
			wantErr: "failed to delete ingredients",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _ := newIngredientSvc(t)
			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			err := svc.DeleteIngredientFromCocktail(tc.ctx, cocktailID, ingredientID)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestGetIngredientsByCocktailID(t *testing.T) {
	const cocktailID = "cocktail-1"

	tests := []struct {
		name         string
		setup        func(repo *mocks.IngredientRepository, ctx context.Context)
		wantErr      string
		expectedData []models.Ingredient
	}{
		{
			name: "success",
			setup: func(repo *mocks.IngredientRepository, ctx context.Context) {
				out := []models.Ingredient{
					{Id: "i1", Name: models.IngredientNomenclature{Name: "Gin"}},
					{Id: "i2", Name: models.IngredientNomenclature{Name: "Tonic"}},
				}
				repo.EXPECT().
					GetIngredientsByCocktailID(ctx, cocktailID).
					Return(out, nil)
			},
			expectedData: []models.Ingredient{
				{Id: "i1", Name: models.IngredientNomenclature{Name: "Gin"}},
				{Id: "i2", Name: models.IngredientNomenclature{Name: "Tonic"}},
			},
		},
		{
			name: "repo error",
			setup: func(repo *mocks.IngredientRepository, ctx context.Context) {
				repo.EXPECT().
					GetIngredientsByCocktailID(ctx, cocktailID).
					Return(nil, errors.New("db error"))
			},
			wantErr: "get ingredients",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo, _ := newIngredientSvc(t)
			ctx := context.Background()

			if tc.setup != nil {
				tc.setup(repo, ctx)
			}

			got, err := svc.GetIngredientsByCocktailID(ctx, cocktailID)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Len(t, got, 0)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedData, got)
		})
	}
}
