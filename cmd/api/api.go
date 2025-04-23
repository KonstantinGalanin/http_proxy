package main

import (
	"database/sql"
	"fmt"
	"net/http"
	"os"

	"github.com/KonstantinGalanin/http_proxy/internal/handlers"
	"github.com/KonstantinGalanin/http_proxy/internal/repository"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var (
	dbHost = os.Getenv("DATABASE_HOST")
	dbPort = os.Getenv("DATABASE_PORT")
	dbUser = os.Getenv("DATABASE_USER")
	dbPass = os.Getenv("DATABASE_PASSWORD")
	dbName = os.Getenv("DATABASE_NAME")

)

func main() {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable", dbHost, dbPort, dbUser, dbPass, dbName)
	db, err := sql.Open("postgres", dsn)
	if err != nil {
		panic(err)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		panic(err)
	}

	repository := repository.New(db)
	handler := handlers.New(repository)

	r := mux.NewRouter()

	r.HandleFunc("/requests", handler.GetRequests).Methods(http.MethodGet)
	r.HandleFunc("/requests/{id}", handler.GetRequestByID).Methods(http.MethodGet)
	r.HandleFunc("/repeat/{id}", handler.RepeatRequest).Methods(http.MethodGet)
	r.HandleFunc("/scan/{id}", handler.ScanRequest).Methods(http.MethodGet)

	http.ListenAndServe(":8000", r)
}
