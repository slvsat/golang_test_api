package server

import (
	"go.api/dblogic"
	"log"
	"net/http"
)

//Creating an instance of Repository
var repository dblogic.Repository = *dblogic.NewRepository()

//Running server on port 8080
func Run(conf dblogic.Config){
	router := NewRouter()
	repository.SetConfig(conf)
	log.Fatal(http.ListenAndServe(":8080", router))
}
