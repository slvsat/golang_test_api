package server

import (
	"net/http"
)

type Route struct {
	Name string
	Method string
	Pattern string
	HandlerFunc http.HandlerFunc
}

type Routes []Route

var routes = Routes{
	Route{
		"Index",
		"GET",
		"/items",
		Index,
	},
	Route{
		"Get data by ID",
		"GET",
		"/items/{itemId}",
		GetDataById,
	},
	Route{
		"Add Data",
		"POST",
		"/items",
		AddData,
	},
	Route{
		"Update Data",
		"PUT",
		"/items/{itemId}",
		UpdateData,
	},
	Route{
		"Delete Data",
		"DELETE",
		"/items/{itemId}",
		DeleteData,
	},
}