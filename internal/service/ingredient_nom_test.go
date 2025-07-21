package service

import (
	"context"
	"errors"
	"nickbar_v2/internal/mocks"
	"nickbar_v2/internal/models/models"
	"nickbar_v2/internal/models/requests"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newIngredientNomSvc(t *testing.T) (*IngredientNomService, *mocks.IngredientNomRepository) {
	repoIngrNom := mocks.NewIngredientNomRepository(t)
	return NewIngredientNomService(repoIngrNom), repoIngrNom
}

func TestCreateIngredientNom(t *testing.T) {
	req := requests.IngredientNomRequest{
		Name:    "Lemon",
		Picture: "lemon.png",
	}

	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context)
		want    models.IngredientNomenclature
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
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				repo.EXPECT().
					CreateIngredientNom(ctx, models.IngredientNomenclature{
						Name:    "Lemon",
						Picture: "lemon.png",
						UserId:  "u1",
					}).
					Return(models.IngredientNomenclature{
						Id:      "nom1",
						Name:    "Lemon",
						Picture: "lemon.png",
						UserId:  "u1",
					}, nil)
			},
			want: models.IngredientNomenclature{
				Id:      "nom1",
				Name:    "Lemon",
				Picture: "lemon.png",
				UserId:  "u1",
			},
		},
		{
			name: "repo error",
			ctx:  ctxWithUser("admin", 1),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				repo.EXPECT().
					CreateIngredientNom(ctx, models.IngredientNomenclature{
						Name:    "Lemon",
						Picture: "lemon.png",
						UserId:  "admin",
					}).
					Return(models.IngredientNomenclature{}, errors.New("db fail"))
			},
			wantErr: "createingredientnom",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			got, err := svc.CreateIngredientNom(tc.ctx, req)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetIngredientNoms(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context)
		want    []models.IngredientNomenclature
		wantErr string
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name:    "not admin → restricted",
			ctx:     ctxWithUser("user", 0),
			wantErr: "restricted access",
		},
		{
			name: "admin success",
			ctx:  ctxWithUser("admin", 1),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				result := []models.IngredientNomenclature{
					{Id: "n1", Name: "Lime"},
					{Id: "n2", Name: "Mint"},
				}
				repo.EXPECT().
					GetIngredientNoms(ctx).
					Return(result, nil)
			},
			want: []models.IngredientNomenclature{
				{Id: "n1", Name: "Lime"},
				{Id: "n2", Name: "Mint"},
			},
		},
		{
			name: "admin repo error",
			ctx:  ctxWithUser("admin", 1),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				repo.EXPECT().
					GetIngredientNoms(ctx).
					Return(nil, errors.New("db fail"))
			},
			wantErr: "getnomenclatures",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			got, err := svc.GetIngredientNoms(tc.ctx)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetUnapprovedIngredientNom(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context)
		want    []models.IngredientNomenclature
		wantErr string
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name:    "not admin",
			ctx:     ctxWithUser("user", 0),
			wantErr: "restricted access",
		},
		{
			name: "admin success",
			ctx:  ctxWithUser("admin", 1),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				result := []models.IngredientNomenclature{
					{Id: "n1", Name: "Cinnamon"},
					{Id: "n2", Name: "Nutmeg"},
				}
				repo.EXPECT().
					GetUnapprovedIngredientNom(ctx).
					Return(result, nil)
			},
			want: []models.IngredientNomenclature{
				{Id: "n1", Name: "Cinnamon"},
				{Id: "n2", Name: "Nutmeg"},
			},
		},
		{
			name: "admin repo error",
			ctx:  ctxWithUser("admin", 1),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				repo.EXPECT().
					GetUnapprovedIngredientNom(ctx).
					Return(nil, errors.New("db fail"))
			},
			wantErr: "getunapprovednomenclatures",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			got, err := svc.GetUnapprovedIngredientNom(tc.ctx)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetApprovedIngredientNom(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context)
		want    []models.IngredientNomenclature
		wantErr string
	}{
		{
			name: "success",
			ctx:  context.Background(),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				result := []models.IngredientNomenclature{
					{Id: "n1", Name: "Vodka", Approved: true},
					{Id: "n2", Name: "Tequila", Approved: true},
				}
				repo.EXPECT().
					GetApprovedIngredientNom(ctx).
					Return(result, nil)
			},
			want: []models.IngredientNomenclature{
				{Id: "n1", Name: "Vodka", Approved: true},
				{Id: "n2", Name: "Tequila", Approved: true},
			},
		},
		{
			name: "repository error",
			ctx:  context.Background(),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context) {
				repo.EXPECT().
					GetApprovedIngredientNom(ctx).
					Return(nil, errors.New("db connection error"))
			},
			wantErr: "getapprovednomenclatures",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			if tc.setup != nil {
				tc.setup(repo, tc.ctx)
			}

			got, err := svc.GetApprovedIngredientNom(tc.ctx)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestGetPersonalIngredientNom(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context, userID string)
		want    []models.IngredientNomenclature
		wantErr string
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			wantErr: "unathorized",
		},
		{
			name: "success",
			ctx:  ctxWithUser("u123", 0),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context, userID string) {
				result := []models.IngredientNomenclature{
					{Id: "n1", Name: "Ice", UserId: userID},
					{Id: "n2", Name: "Lime", UserId: userID},
				}
				repo.EXPECT().
					GetPersonalIngredientNom(ctx, userID).
					Return(result, nil)
			},
			want: []models.IngredientNomenclature{
				{Id: "n1", Name: "Ice", UserId: "u123"},
				{Id: "n2", Name: "Lime", UserId: "u123"},
			},
		},
		{
			name: "repository error",
			ctx:  ctxWithUser("u456", 0),
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context, userID string) {
				repo.EXPECT().
					GetPersonalIngredientNom(ctx, userID).
					Return(nil, errors.New("query error"))
			},
			wantErr: "getpersonalnoms",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			user, _ := models.GetUserFromContext(tc.ctx)

			if tc.setup != nil {
				tc.setup(repo, tc.ctx, user.Id)
			}

			got, err := svc.GetPersonalIngredientNom(tc.ctx)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
				assert.Empty(t, got)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.want, got)
		})
	}
}

func TestUpdateIngredientNomApprovedStatus(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		request []requests.IngredientNomUpdateRequest
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context, expected []models.IngredientNomenclature)
		wantErr string
	}{
		{
			name:    "unauthorized",
			ctx:     context.Background(),
			request: nil,
			wantErr: "unathorized",
		},
		{
			name:    "restricted access",
			ctx:     ctxWithUser("user1", 0),
			request: []requests.IngredientNomUpdateRequest{{Id: "n1", Approved: true}},
			wantErr: "restricted access",
		},
		{
			name: "success",
			ctx:  ctxWithUser("admin", 1),
			request: []requests.IngredientNomUpdateRequest{
				{Id: "n1", Approved: true, Picture: "img1"},
				{Id: "n2", Approved: false, Picture: "img2"},
			},
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context, expected []models.IngredientNomenclature) {
				repo.EXPECT().
					UpdateIngredientNomsApprovedStatus(ctx, expected).
					Return(nil)
			},
		},
		{
			name: "repo error",
			ctx:  ctxWithUser("admin", 1),
			request: []requests.IngredientNomUpdateRequest{
				{Id: "n1", Approved: true, Picture: "img1"},
			},
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context, expected []models.IngredientNomenclature) {
				repo.EXPECT().
					UpdateIngredientNomsApprovedStatus(ctx, expected).
					Return(errors.New("db fail"))
			},
			wantErr: "failed to update nomenclatures",
		},
	}

	for _, tc := range tests {
		tc := tc // capture range variable
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			var expected []models.IngredientNomenclature
			for _, r := range tc.request {
				expected = append(expected, models.IngredientNomenclature{
					Id:       r.Id,
					Picture:  r.Picture,
					Approved: r.Approved,
				})
			}

			if tc.setup != nil {
				tc.setup(repo, tc.ctx, expected)
			}

			err := svc.UpdateIngredientNomApprovedStatus(tc.ctx, tc.request)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestDeleteIngredientNom(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		id      string
		setup   func(repo *mocks.IngredientNomRepository, ctx context.Context, id string)
		wantErr string
	}{
		{
			name:    "unauthorized user",
			ctx:     context.Background(),
			id:      "nom1",
			wantErr: "unathorized",
		},
		{
			name:    "restricted access for non-admin",
			ctx:     ctxWithUser("user1", 0),
			id:      "nom1",
			wantErr: "restricted access",
		},
		{
			name: "successful delete by admin",
			ctx:  ctxWithUser("admin", 1),
			id:   "nom1",
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context, id string) {
				repo.EXPECT().
					DeleteIngredientNom(ctx, id).
					Return(nil)
			},
		},
		{
			name: "repo error",
			ctx:  ctxWithUser("admin", 1),
			id:   "nom1",
			setup: func(repo *mocks.IngredientNomRepository, ctx context.Context, id string) {
				repo.EXPECT().
					DeleteIngredientNom(ctx, id).
					Return(errors.New("db failure"))
			},
			wantErr: "failed to delete nomenclature",
		},
	}

	for _, tc := range tests {
		tc := tc // захват переменной
		t.Run(tc.name, func(t *testing.T) {
			svc, repo := newIngredientNomSvc(t)

			if tc.setup != nil {
				tc.setup(repo, tc.ctx, tc.id)
			}

			err := svc.DeleteIngredientNom(tc.ctx, tc.id)

			if tc.wantErr != "" {
				assert.ErrorContains(t, err, tc.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
