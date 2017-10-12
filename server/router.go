package server

import (
	"github.com/gorilla/mux"
)

//Creating new Router with routes written in routes.go
func NewRouter() *mux.Router{
	router := mux.NewRouter().StrictSlash(true)
	for _, route := range routes{
		router.
		Methods(route.Method).
			Path(route.Pattern).
			Name(route.Name).
			Handler(route.HandlerFunc)
	}
	return router
}