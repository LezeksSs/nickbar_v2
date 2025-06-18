package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	// ---------- флаги ----------
	cmd := flag.String("cmd", "up", "up | down | steps | force | version")
	n := flag.Int("n", 0, "кол-во шагов для down/steps (0 = все)")
	fpath := flag.String("path", "./migrations", "путь к SQL-миграциям")
	force := flag.Int("v", 0, "force: установить версию")
	flag.Parse()

	// ---------- DSN из env или значений по умолчанию ----------
	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		getenv("POSTGRES_USER", "postgres"),
		getenv("POSTGRES_PASSWORD", "postgres"),
		getenv("POSTGRES_HOST", "localhost"),
		getenv("POSTGRES_PORT", "5432"),
		getenv("POSTGRES_DB", "postgres"),
	)

	// ---------- инициализация migrate ----------
	db, err := sql.Open("pgx", dsn)
	check(err)
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	check(err)

	m, err := migrate.NewWithDatabaseInstance(
		"file://"+*fpath, // источник SQL-файлов
		"postgres", driver,
	)
	check(err)

	// ---------- выполнение ----------
	switch *cmd {
	case "up":
		err = m.Up()

	case "down":
		if *n == 0 {
			err = m.Down()
		} else {
			err = m.Steps(-*n)
		}

	case "steps":
		if *n == 0 {
			log.Fatal("для steps укажите -n≠0")
		}
		err = m.Steps(*n)

	case "force":
		err = m.Force(*force)

	case "version":
		v, dirty, err2 := m.Version()
		check(err2)
		log.Printf("текущая версия: %d (dirty=%v)\n", v, dirty)
		return

	default:
		log.Fatalf("неизвестная команда: %s", *cmd)
	}

	if err != nil && err != migrate.ErrNoChange {
		log.Fatalf("ошибка миграции: %v", err)
	}

	log.Println("✅ готово")
}

func getenv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func check(err error) {
	if err != nil {
		log.Fatal(err)
	}
}
