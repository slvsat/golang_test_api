package server

import (
	http "net/http"
	"encoding/json"
	"log"
	"io/ioutil"
	"io"
	"strings"
	"github.com/gorilla/mux"
	"gopkg.in/mgo.v2/bson"
	"go.api/dblogic"
)


func WriteHeader(writer http.ResponseWriter){
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	return
}

func Index(writer http.ResponseWriter, request *http.Request) {
	queryString := request.URL.Query().Get("q")
	log.Println("URL : ", request.URL.RequestURI())
	data, _ := repository.GetData(request.URL.RequestURI(), queryString)
	WriteHeader(writer)
	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
	return
}

func GetDataById(writer http.ResponseWriter, request *http.Request){
	vars := mux.Vars(request)
	id := vars["itemId"]
	log.Println("URL : ", request.URL.RequestURI())
	data, _ := repository.GetDataById(request.URL.RequestURI(), id)
	WriteHeader(writer)
	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
	return
}

func AddData(writer http.ResponseWriter, request *http.Request){
	var data dblogic.Data
	body, err := ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
	if err != nil {
		log.Fatalln("Error data: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := request.Body.Close(); err != nil {
		log.Fatalln("Error AddData: ", err)
	}
	if err := json.Unmarshal(body, &data); err != nil {
		writer.WriteHeader(422)
		if err := json.NewEncoder(writer).Encode(err); err != nil {
			log.Fatalln("Error addData unmarshalling data: ", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	success, _ := repository.AddData(data)
	if success == "0" {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	writer.Write([]byte(success))
	writer.WriteHeader(http.StatusCreated)
	return
}

func UpdateData(writer http.ResponseWriter, request *http.Request){
	var data dblogic.Data
	body, err := ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
	if err != nil {
		log.Fatalln("Error Update Data: ", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	if err := request.Body.Close(); err != nil {
		log.Fatalln("Error update data (Body.Close()): ", err)
	}
	if err := json.Unmarshal(body, &data); err != nil {
		WriteHeader(writer)
		writer.WriteHeader(422)
		if err := json.NewEncoder(writer).Encode(err); err != nil {
			log.Fatalln("Erorr updateData unmarshalling data ", err)
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
	vars := mux.Vars(request)
	data.Id = bson.ObjectIdHex(vars["itemId"])
	success, _ := repository.UpdateData(data)
	if !success {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	WriteHeader(writer)
	writer.WriteHeader(http.StatusOK)
	return
}

func DeleteData (writer http.ResponseWriter, request *http.Request){
	vars := mux.Vars(request)
	itemId := vars["itemId"]
	if err := repository.DeleteData(itemId); err != "" {
		if strings.Contains(err, "404") {
			writer.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(err, "500"){
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	WriteHeader(writer)
	writer.WriteHeader(http.StatusOK)
	return
}