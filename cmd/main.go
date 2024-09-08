package main

import (
	"fmt"

	"github.com/wbrijesh/identity/internal/database"
	"github.com/wbrijesh/identity/internal/server"
)

func init() {
	dbService := database.New()
	dbService.RunMigrations()
}

func main() {
	server := server.NewServer()
	err := server.ListenAndServe()
	if err != nil {
		panic(fmt.Sprintf("cannot start server: %s", err))
	}
}
