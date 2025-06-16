package main

import (
	"fmt"
	"log"
	"nickbar_v2/internal/config"
)

func main() {
	// TODO: init config - viper/cleanenv
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("error in config: %s", err)
	}
	fmt.Println(cfg)
	// TODO: init logger - log/slog

	// TODO: init storage - PostgreSQL

	// TODO: init router - gorilla/mux

	// TODO: run server

}
