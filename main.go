package main

import (
	"log"
	"os"

	"go_final_project/internal/db"
	"go_final_project/internal/server"
)

func main() {
	logger := log.New(os.Stdout, "[INFO] ", log.LstdFlags)

	dbFile := os.Getenv("TODO_DBFILE")
	if dbFile == "" {
		dbFile = "scheduler.db"
	}
	if err := db.Init(dbFile); err != nil {
		logger.Fatalf("failed to init db: %v", err)
	}

	port := os.Getenv("TODO_PORT")
	if port == "" {
		port = "7540"
		logger.Println("No port specified, using default port " + port)
	}

	srv := server.NewServer(logger, port)
	if err := srv.Run(); err != nil {
		logger.Fatal(err)
	}
}
