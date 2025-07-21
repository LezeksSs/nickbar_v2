package service

import (
	"errors"
	"nickbar_v2/internal/mocks"
	"nickbar_v2/internal/models/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

// ───────────────── helpers ────────────────────────────────────────────────

func newAuthSvc(t *testing.T) (*AuthService,
	*mocks.AuthRepository, *mocks.UserRepository) {

	authRepo := mocks.NewAuthRepository(t)
	userRepo := mocks.NewUserRepository(t)
	return NewAuthService(authRepo, userRepo), authRepo, userRepo
}

// ───────────────── AuthUser tests ─────────────────────────────────────────

func TestAuthUser(t *testing.T) {
	tests := []struct {
		name       string
		token      string
		setup      func(a *mocks.AuthRepository, u *mocks.UserRepository)
		wantUser   models.User
		wantErrMsg string
	}{
		{
			name:  "success",
			token: "good",
			setup: func(a *mocks.AuthRepository, u *mocks.UserRepository) {
				a.EXPECT().
					ValidateToken("good").
					Return(true, "nick", nil)

				u.EXPECT().
					GetUser("nick").
					Return(models.User{Id: "u1", Nickname: "nick", Role: 0}, nil)
			},
			wantUser: models.User{Id: "u1", Nickname: "nick", Role: 0},
		},
		{
			name:  "invalid token",
			token: "bad",
			setup: func(a *mocks.AuthRepository, u *mocks.UserRepository) {
				a.EXPECT().
					ValidateToken("bad").
					Return(false, "", nil)
			},
			wantErrMsg: "Unauthorized",
		},
		{
			name:  "validate token error",
			token: "err",
			setup: func(a *mocks.AuthRepository, _ *mocks.UserRepository) {
				a.EXPECT().
					ValidateToken("err").
					Return(false, "", errors.New("redis down"))
			},
			wantErrMsg: "validatetoken",
		},
		{
			name:  "get user error",
			token: "good2",
			setup: func(a *mocks.AuthRepository, u *mocks.UserRepository) {
				a.EXPECT().
					ValidateToken("good2").
					Return(true, "nick2", nil)

				u.EXPECT().
					GetUser("nick2").
					Return(models.User{}, errors.New("not found"))
			},
			wantErrMsg: "getuser",
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			svc, authRepo, userRepo := newAuthSvc(t)
			tc.setup(authRepo, userRepo)

			user, err := svc.AuthUser(tc.token) // tc.token = токен строка
			if tc.wantErrMsg != "" {
				assert.ErrorContains(t, err, tc.wantErrMsg)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tc.wantUser, user)
		})
	}
}

// ───────────────── CheckAccess tests ──────────────────────────────────────

func TestCheckAccess(t *testing.T) {
	svc, _, _ := newAuthSvc(t) // репо не нужны

	admin := models.User{Id: "adm", Role: 1}
	user := models.User{Id: "u1", Role: 0}

	tests := []struct {
		name     string
		route    string
		method   string
		user     models.User
		expected bool
	}{
		{"admin POST /cocktails", "/cocktails", "POST", admin, true},
		{"user DELETE /cocktails (forbidden)", "/cocktails", "DELETE", user, false},
		{"unknown route", "/unknown", "GET", admin, false},
		{"unknown method", "/cocktails", "PATCH", admin, false},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			ok := svc.CheckAccess(tc.route, tc.method, tc.user)
			assert.Equal(t, tc.expected, ok)
		})
	}
}
