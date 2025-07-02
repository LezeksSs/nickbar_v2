package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"nickbar_v2/internal/config"
	"nickbar_v2/internal/storage/pgsql"
	"os"
)

const (
	envLocal = "local"
	envProd  = "prod"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error in config: %s", err)
	}

	// TODO: Delete later
	fmt.Println(cfg)

	// APP_ENV as in config.go - local/prod
	log := setupLogger(os.Getenv("APP_ENV"))

	log.Info("starting nickbar")
	log.Debug("debug messages are enabled")

	// TODO: init storage - PostgreSQL
	_, err = pgsql.NewPgxPool(ctx)

	// TODO: init router - gorilla/mux

	// TODO: run server

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
