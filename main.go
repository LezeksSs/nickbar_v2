package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"nickbar_v2/internal/config"
	"nickbar_v2/internal/controller"
	"nickbar_v2/internal/repository"
	"nickbar_v2/internal/service"
	"nickbar_v2/internal/storage/pgsql"
	"os"

	"github.com/google/uuid"
	"github.com/gorilla/mux"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

var tokenizer func() (string, string, error) = func() (string, string, error) {
	accessToken := uuid.NewString()
	refreshToken := uuid.NewString()
	return accessToken, refreshToken, nil
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error in config: %s", err)
	}

	// TODO: Delete later
	fmt.Println(cfg)

	// APP_ENV authService in config.go - local/prod
	log := setupLogger(os.Getenv("APP_ENV"))

	log.Info("starting nickbar")
	log.Debug("debug messages are enabled")

	// TODO: init storage - PostgreSQL
	pool, err := pgsql.NewPgxPool(ctx)

	redisRepo, err := repository.NewRedisAuthRepository(repository.RedisAuthParams{
		Addr:         cfg.Redis.Address,
		WriteTimeout: cfg.Redis.WriteTimeout,
		ReadTimeout:  cfg.Redis.ReadTimeout,
		TokenExpire:  cfg.Redis.TokenExpire,
		Tokenizer:    tokenizer,
	})

	cocktailRepo := repository.NewCocktailRepository(pool)
	ingrRepo := repository.NewIngredientRepository(pool)
	ingrNomRepo := repository.NewIngredientNomenclatureRepository(pool)
	tagRepo := repository.NewTagRepository(pool)
	userRepo := repository.NewUserRepository(pool)
	authRepo, _ := repository.NewRedisAuthRepository(repository.RedisAuthParams{})

	// TODO: init router - gorilla/mux

	authService := service.NewAuthService(redisRepo, userRepo)
	cocktailService := service.NewCocktailService(cocktailRepo, ingrRepo, tagRepo, ingrNomRepo)
	ingrService := service.NewIngredientService(ingrRepo, ingrNomRepo)
	ingrNomService := service.NewIngredientNomService(ingrNomRepo)
	tagService := service.NewTagService(tagRepo)
	userService := service.NewUserService(userRepo, authRepo)

	am := controller.NewAuthMiddleware(log, authService)
	cc := controller.NewCocktailController(log, cocktailService)
	ic := controller.NewIngredientController(log, ingrService)
	inc := controller.NewIngredientNomController(log, ingrNomService)
	tc := controller.NewTagController(log, tagService)
	ac := controller.NewAuthController(log, userService)

	r := setupRouter(am.Logging)

	// TODO: run server

	r.HandleFunc("/login", ac.LoginUser).Methods("POST")
	r.HandleFunc("/register", ac.RegisterUser).Methods("POST")

	s := r.PathPrefix("/").Subrouter()
	s.HandleFunc("/cocktails", cc.CreateCocktail).Methods("POST")
	s.HandleFunc("/cocktails", cc.GetCocktails).Methods("GET")
	s.HandleFunc("/cocktails/search", cc.SearchCocktails).Methods("GET")
	s.HandleFunc("/cocktails", cc.DeleteCocktail).Methods("DELETE")

	addr := fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.Port)
	log.Info("Starting server", slog.Any("addr", addr))

	if err := http.ListenAndServe(addr, r); err != nil {
		log.Error("server failed: %v", err)
	}

	// TODO: implement ingredient quantity checker
	// TODO: implement event/notify feature   client -> server -> kitchen api -> waitress api (notification)

	/*
		Кухня будет заполнять/добавлять коктейли и добавлять/увеличивать ингредиенты.
		У кухни будет дополнительный функционал для осуществления заказов
	*/

	/*
		Юзеры (кастомер-интерфейс) смогут добавлять коктейли и они будут рассматриваться на одобрение кухней (отдельное окно у кухни для просмотра предложений с добавлением коктейля в меню)
	*/

	/*
		АПИ для оффициантов (кастомер-интерфейс) будет реализовывать прием заказов от клиентов и передачу ее на кухню, кухня после выполнения заказа (сделала коктейль) делает нотификацию
		для оффицианта, что заказ N готов к передаче клиенту.
	*/

	/*
		User/Admin/Waitress будут акторами

		Основной сервис (этот) будет выполнять функционал аутентификации, создание юзеров, добавление новых коктейлей (меню),
		апрув пользовательских коктейлей, тегов, ингредиентов (номенклатуры) админом

		Сервис официантов (Waitress API) будет отправлять запрос заказов на кухню (Кафка - publish/subscribe OrderPlaced)

		Сервис кухни (Kitchet API) принимает запросы от официантов и отправляет нотифай (Кафка - publish/subscribe OrderReady)

		API Gateway будет перенаправлять запросы в Кафку, откуда уже будет вытягиваться данные подписчиками
	*/
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger
	switch env {
	case envLocal:
		log = slog.New(
			slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}),
		)
	case envProd:
		log = slog.New(
			slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}),
		)
	}

	return log
}

func setupRouter(mws ...mux.MiddlewareFunc) *mux.Router {
	r := mux.NewRouter()

	r.Use(mws...)

	return r
}
