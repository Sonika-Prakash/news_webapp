package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"webapp/base"

	_ "github.com/lib/pq"
	"github.com/upper/db/v4"
	"github.com/upper/db/v4/adapter/postgresql"
)

// DSN is the connection string to postgres DB
const (
	DSN = "postgres://soprakash:postgres@localhost/newswebapp?sslmode=disable"
)

func runMigration(db db.Session) error {
	script, err := os.ReadFile("./migrations/tables.sql")
	if err != nil {
		return err
	}
	fmt.Println("Starting DB migration")
	_, err = db.SQL().Exec(string(script))
	if err != nil {
		return err
	}
	fmt.Println("DB migration done!")
	return nil
}

func main() {
	migrate := flag.Bool("migrate", false, "For DB migration")
	flag.Parse()

	db := base.OpenDB(DSN)
	defer db.Close()
	upper, err := postgresql.New(db) // upper is used just to provide nice methods to run operations on db
	if err != nil {
		log.Fatalln("Error while creating an upper wrapper for DB instance", err)
	}
	defer upper.Close()

	if *migrate {
		err = runMigration(upper)
		if err != nil {
			log.Fatalln("Error while database migration", err)
		}
	}

	app := base.GetApplicationInstance("NewsWebApp", "localhost", "8080", db, upper)
	h := base.MakeHTTPHandler(app)
	srv := app.GetServer(h)

	errs := make(chan error)
	go func() {
		errs <- srv.ListenAndServe()
	}()

	go func() {
		app.CatchInterruptions(errs)
	}()

	err = <-errs
	app.GracefulShutdown(srv, err)
}
