package server

import (
	"testing"
	"go.api/dblogic"
	"net/http"
	"log"
	"net/http/httptest"
	"bytes"
	"github.com/gorilla/mux"
)

//TESTING

var rep dblogic.Repository = *dblogic.NewRepository()

func TestGetRequest(t *testing.T){
	rep.ClearTable()

	req, err := http.NewRequest("GET", "/items", nil)
	if err != nil {
		log.Println("Request error! ", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Index)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK{
		t.Errorf("Returned wrong status code, got %v want %v ", status, http.StatusOK)
	}

	expected := "[]"
	if rr.Body.String() != expected {
		t.Errorf("Returned unexpected body, got %v want %v", rr.Body, expected)
	}
}

var Id string

func TestPostRequest(t *testing.T) {
	reqBody := `{"name":"Sattar", "data_itself": "Hello"}`
	req, err := http.NewRequest("POST", "/items", bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		t.Fatalf("Cannot create request! %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(AddData)
	handler.ServeHTTP(rr, req)

	if rr.Body == nil {
		t.Errorf("Returned nil, expected ID")
	}
	Id = rr.Body.String()
}

func TestPutRequest(t *testing.T){
	reqBody := `{"name": "NewOne", "data_itself": "SecondOne"}`
	req, err := http.NewRequest("PUT", "/items/" + Id, bytes.NewBuffer([]byte(reqBody)))
	if err != nil {
		t.Fatalf("Cannot create update request! %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	handler := mux.NewRouter()
	handler.HandleFunc("/items/{itemId}", UpdateData)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Expected $v got $v", http.StatusOK, status)
	}
}

func TestDeleteData(t *testing.T) {
	req, err := http.NewRequest("DELETE", "/items/" + Id, nil)
	if err != nil {
		t.Fatalf("Cannot create delete request! %v", err)
	}
	rr := httptest.NewRecorder()
	handler := mux.NewRouter()
	handler.HandleFunc("/items/{itemId}", DeleteData)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Error deleting item! Expected %v, got %v ", http.StatusOK, status)
	}
}

//BENCHMARKING

func BenchmarkIndex(b *testing.B) {
	rep.ClearTable()
	req, err := http.NewRequest("GET", "/items", nil)
	if err != nil {
		log.Println("Request error! ", err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(Index)

	handler.ServeHTTP(rr, req)
	if status := rr.Code; status != http.StatusOK{
		b.Errorf("Returned wrong status code, got %v want %v ", status, http.StatusOK)
	}

	if rr.Body.String() != "[]" {
		b.Errorf("Returned unexpected body, got %v wanted %v", rr.Body, "[]")
	}
}

var Ids []string

func BenchmarkAddData(b *testing.B) {
	for i := 0; i<b.N; i++{
		reqBody := `{"name":"Sattar", "data_itself": "Hello"}`
		req, err := http.NewRequest("POST", "/items", bytes.NewBuffer([]byte(reqBody)))
		if err != nil {
			b.Fatalf("Cannot create request! %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := http.HandlerFunc(AddData)
		handler.ServeHTTP(rr, req)

		//log.Println(rr.Body)
		if rr.Body == nil {
			b.Errorf("Returned nil, expected ID")
		}
		Ids = append(Ids, rr.Body.String())
	}
}

func BenchmarkUpdateData(b *testing.B) {
	for _, element := range Ids {
		reqBody := `{"name": "NewOne", "data_itself": "SecondOne"}`
		req, err := http.NewRequest("PUT", "/items/" + element, bytes.NewBuffer([]byte(reqBody)))
		if err != nil {
			b.Fatalf("Cannot create update request! %v", err)
		}
		req.Header.Set("Content-Type", "application/json")

		rr := httptest.NewRecorder()
		handler := mux.NewRouter()
		handler.HandleFunc("/items/{itemId}", UpdateData)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			b.Errorf("Expected $v got $v", http.StatusOK, status)
		}
	}
}

func BenchmarkDeleteData(b *testing.B) {
	for _, element := range Ids {
		req, err := http.NewRequest("DELETE", "/items/" + element, nil)
		if err != nil {
			b.Fatalf("Cannot create delete request! %v", err)
		}
		rr := httptest.NewRecorder()
		handler := mux.NewRouter()
		handler.HandleFunc("/items/{itemId}", DeleteData)
		handler.ServeHTTP(rr, req)

		if status := rr.Code; status != http.StatusOK {
			b.Errorf("Error deleting item! Expected %v, got %v ", http.StatusOK, status)
		}
	}
}