package models

import "context"

// Ключ для хранения данных о пользователе в контексте
type contextKey string

const userKey contextKey = "user"

// Функция добавления User в Context
func WithUser(ctx context.Context, user User) context.Context {
	return context.WithValue(ctx, userKey, user)
}

// Функция извлечения User из Context
func GetUserFromContext(ctx context.Context) (User, bool) {
	user, ok := ctx.Value(userKey).(User)
	return user, ok
}
