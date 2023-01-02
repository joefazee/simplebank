package main

import (
	"database/sql"
	"log"

	"github.com/joefazee/simplebank/api"
	db "github.com/joefazee/simplebank/db/sqlc"
	"github.com/joefazee/simplebank/util"
	_ "github.com/lib/pq"
)

func main() {

	config, err := util.LoadCondfig(".")

	if err != nil {
		log.Fatal(err)
	}

	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal(err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("unable to create server")
	}

	err = server.Start(config.ServerAddress)

	if err != nil {
		log.Fatal(err)
	}
}
