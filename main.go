package main

import (
	"database/sql"
	"log"
	"simplebank/api"
	db "simplebank/db/sqlc"
	"simplebank/util"

	_ "github.com/lib/pq"
	_ "simplebank/docs"
)

// @title        Simple Bank API
// @version      1.0
// @description  This is a simple bank API.

// @host      localhost:8080
// @BasePath  /

// @securitydefinitions.apikey  ApiKeyAuth
// @in                          header
// @name                        Authorization

func main() {
	config, err := util.LoadConfig(".")
	if err != nil {
		log.Fatal("cannot load config:", err)
	}
	conn, err := sql.Open(config.DBDriver, config.DBSource)
	if err != nil {
		log.Fatal("cannot connect to db:", err)
	}

	store := db.NewStore(conn)
	server, err := api.NewServer(config, store)
	if err != nil {
		log.Fatal("cannot create server:", err)
	}

	err = server.Start(config.ServerAddress)
	if err != nil {
		log.Fatal("cannot run server:", err)
	}
}
