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

//Writing content type and access to the header
func WriteHeader(writer http.ResponseWriter){
	writer.Header().Set("Content-Type", "application/json; charset=UTF-8")
	writer.Header().Set("Access-Control-Allow-Origin", "*")
	return
}

//Taking url query if exist and returning a data from GetData function
func Index(writer http.ResponseWriter, request *http.Request) {
	queryString := request.URL.Query().Get("q")
	//log.Println("URL : ", request.URL.RequestURI())
	data, err := repository.GetData(request.URL.RequestURI(), queryString)
	WriteHeader(writer)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(data)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}

//Taking ID specified in url and returning a value corresponding to this ID
func GetDataById(writer http.ResponseWriter, request *http.Request){
	vars := mux.Vars(request)
	id := vars["itemId"]
	log.Println("URL : ", request.URL.RequestURI())
	data, err := repository.GetDataById(request.URL.RequestURI(), id)
	WriteHeader(writer)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		writer.Write(data)
		return
	}
	writer.WriteHeader(http.StatusOK)
	writer.Write(data)
}

//Adding data to the Database, returning added Object Id
func AddData(writer http.ResponseWriter, request *http.Request){
	var data dblogic.Data
	body, err := ioutil.ReadAll(io.LimitReader(request.Body, 1048576))
	if err != nil {
		log.Println("Error data: ", err)
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
}

//Updating data by specified Id, returning status of the Request (http status)
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
	err = repository.UpdateData(data)
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}
	WriteHeader(writer)
	writer.WriteHeader(http.StatusOK)
}

//Delete data by specified Id, returning status of removing (http status)
func DeleteData (writer http.ResponseWriter, request *http.Request){
	vars := mux.Vars(request)
	itemId := vars["itemId"]
	if outString, err := repository.DeleteData(itemId); err != nil {
		if strings.Contains(outString, "404") {
			writer.WriteHeader(http.StatusNotFound)
		} else if strings.Contains(outString, "500"){
			writer.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	WriteHeader(writer)
	writer.WriteHeader(http.StatusOK)
}