package repository

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type UidProducer func() (string, string, error)

type RedisAuthRepository struct {
	conn       redis.Client
	wrTmout    time.Duration
	rdTmout    time.Duration
	tknExpTime time.Duration
	tokenizer  UidProducer
}

type TokenPayload struct {
	AccessToken  string
	RefreshToken string
}

type RedisAuthParams struct {
	Addr         string
	Password     string
	DB           int
	WriteTimeout time.Duration
	ReadTimeout  time.Duration
	TokenExpire  time.Duration
	Tokenizer    UidProducer
}

type AuthRepository interface {
	ProduceToken(nickname string) (TokenPayload, error)
	ValidateToken(accToken string) (bool, string, error)
	RefreshTokens(refrToken string) error
}

func NewRedisAuthRepository(params RedisAuthParams) (AuthRepository, error) {
	// Валидация входящих параметров
	if params.Addr == "" {
		return nil, fmt.Errorf("address cannot be empty")
	}
	if params.WriteTimeout <= 0 {
		return nil, fmt.Errorf("write timeout must be positive")
	}
	if params.ReadTimeout <= 0 {
		return nil, fmt.Errorf("read timeout must be positive")
	}
	if params.TokenExpire <= 0 {
		return nil, fmt.Errorf("token expiration time must be positive")
	}
	if params.Tokenizer == nil {
		return nil, fmt.Errorf("tokenizer cannot be nil")
	}

	// Инициализация клиента Redis
	client := redis.NewClient(&redis.Options{
		Addr:     params.Addr,
		Password: params.Password, // оставляем пустым, если пароля нет
		DB:       params.DB,       // используем по умолчанию 0
	})

	// Проверка подключения к Redis
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %v", err)
	}

	// Создание и возврат экземпляра RedisAuthRepository
	return &RedisAuthRepository{
		conn:       *client,
		wrTmout:    params.WriteTimeout,
		rdTmout:    params.ReadTimeout,
		tknExpTime: params.TokenExpire,
		tokenizer:  params.Tokenizer,
	}, nil
}

func (r *RedisAuthRepository) ProduceToken(nickname string) (TokenPayload, error) {
	//accessToken := uuid.NewString()
	//refreshToken := uuid.NewString()
	accessToken, refreshToken, _ := r.tokenizer()
	ctx, cncl := context.WithTimeout(context.Background(), r.wrTmout)
	defer cncl()
	accSt := r.conn.Set(ctx, accessToken, nickname, r.tknExpTime)
	if accSt.Err() != nil {
		return TokenPayload{}, accSt.Err()
	}
	rfshSt := r.conn.Set(ctx, refreshToken, accessToken+"/"+nickname, r.tknExpTime*2)
	if rfshSt.Err() != nil {
		return TokenPayload{}, rfshSt.Err()
	}

	return TokenPayload{AccessToken: accessToken, RefreshToken: refreshToken}, nil
}

func (r *RedisAuthRepository) ValidateToken(accToken string) (bool, string, error) {
	ctx, cncl := context.WithTimeout(context.Background(), r.rdTmout)
	defer cncl()

	val, err := r.conn.Get(ctx, accToken).Result()
	if err == redis.Nil {
		return false, "", nil
	}
	if err != nil {
		return false, "", err
	}

	return val != "", val, nil
}

func (r *RedisAuthRepository) RefreshTokens(refrToken string) error {
	ctx, cncl := context.WithTimeout(context.Background(), r.wrTmout)
	defer cncl()

	accessToken, err := r.conn.Get(ctx, refrToken).Result()
	if err != nil {
		return err
	}
	splittedAToken := strings.Split(accessToken, "/")

	accTkn, _, _ := r.tokenizer()

	accSt := r.conn.Set(ctx, accTkn, splittedAToken[1], r.tknExpTime)
	if accSt.Err() != nil {
		return accSt.Err()
	}

	_, err2 := r.conn.Expire(ctx, refrToken, r.tknExpTime*2).Result()
	if err2 != nil {
		return err2
	}

	return nil
}
