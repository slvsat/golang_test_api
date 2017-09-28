package server

import (
	"go.api/dblogic"
	"log"
	"net/http"
)
var repository dblogic.Repository = dblogic.Repository{}

func Run(conf dblogic.Config){
	router := NewRouter()
	repository.SetConfig(conf)
	log.Fatal(http.ListenAndServe(":8080", router))
}
