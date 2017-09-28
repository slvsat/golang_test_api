package dblogic

import (
	"fmt"
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/aerospike/aerospike-client-go"
	"encoding/json"
	"strconv"
)
type Repository struct{}

var Conf Config

func (r Repository) SetConfig(config Config){
	Conf = config
}

func connClient() (*aerospike.Client, *aerospike.WritePolicy){
	port, _ := strconv.Atoi(Conf.AerospikePort)
	//log.Println(port)
	//log.Println(Conf.AerospikeHost)
	client, err := aerospike.NewClient(Conf.AerospikeHost, port)
	if err != nil {
		log.Println("Error while connection to aerospike", err)
		panic(err)
	}
	policy := aerospike.NewWritePolicy(0, 10)
	return client, policy
}

func writeDataToAerospike(client *aerospike.Client, key *aerospike.Key, policy *aerospike.WritePolicy, data []Data) bool {
	dataToWrite, _ := json.Marshal(data)
	bins := aerospike.BinMap{
		"bin1" : string(dataToWrite),
	}
	err := client.Put(policy, key, bins)
	if err != nil {
		return false
	}
	return true
}

func getFromAerospike(client *aerospike.Client, key *aerospike.Key) string{
	rec, err := client.Get(nil, key)
	if err != nil {
		log.Println("Error while getting data from aerospike", err)
		panic(err)
	}
	return rec.Bins["bin1"].(string)

}

func parseQuery(q string) bson.M{
	if q != "" {
		outQuery := bson.M{}
		err := bson.UnmarshalJSON([]byte(q), &outQuery)
		if err != nil{
			log.Println("Error while UnmarshalingJSON ", err)
			panic(err)
		}
		log.Println(outQuery)
		return outQuery
	}
	return nil
}

func (r Repository) GetDataById(url string, id string) []byte{
	client, policy := connClient()

	key, err := aerospike.NewKey("test", "aerospike", url)
	if err != nil {
		log.Println("Error while creating a key (get data by id) ", err)
		panic(err)
	}

	exist, err := client.Exists(nil, key)
	if err != nil {
		panic(err)
	}
	result := make([]Data, 0)
	if exist == false{
		session, err := mgo.Dial(Conf.MongoDBhost)
		if err != nil {
			log.Println("Failed to establish connection to mongodb", err)
		}
		defer session.Close()
		c := session.DB(Conf.MongoDBname).C(Conf.MongoDBdocname)
		if err := c.FindId(bson.ObjectIdHex(id)).All(&result); err != nil {
			panic(err)
		}
		writeDataToAerospike(client, key, policy, result)
	}else {
		return []byte(getFromAerospike(client, key))
	}
	output, _ := json.Marshal(result)
	return output
}

func (r Repository) GetData(url string, query string) []byte {
	client, policy := connClient()

	key, err := aerospike.NewKey("test", "aerospike", url)
	if err != nil {
		log.Println("Error while creating a key (aerospike)", err)
		panic(err)
	}

	exist, err := client.Exists(nil, key)
	if err != nil {
		panic(err)
	}
	results := make([]Data, 0)
	if exist == false {
		session, err := mgo.Dial(Conf.MongoDBhost)
		if err != nil {
			log.Println("Failed to establish connection to Mongo server:", err)
		}
		defer session.Close()
		c := session.DB(Conf.MongoDBname).C(Conf.MongoDBdocname)
		if err := c.Find(parseQuery(query)).All(&results); err != nil {
			log.Println("Failed to write results:", err)
		}
		writeDataToAerospike(client, key, policy, results)
	}else {
		return []byte(getFromAerospike(client, key))
	}
	output, _ := json.Marshal(results)
	return output
}


func (r Repository) AddData(data Data) string{
	session, err := mgo.Dial(Conf.MongoDBhost)
	defer session.Close()
	data.Id = bson.NewObjectId()
	session.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).Insert(data)
	if err != nil {
		log.Fatal(err)
		return "Cannot add item"
	}
	return data.Id.String()
}

func (r Repository) UpdateData(data Data) bool {
	session, err := mgo.Dial(Conf.MongoDBhost)
	defer session.Close()
	session.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).UpdateId(data.Id, data)
	if err != nil {
		log.Fatal(err)
		return false
	}
	return true
}


func (r Repository) DeleteData(id string) string{
	session, err := mgo.Dial(Conf.MongoDBhost)
	if err != nil{
		fmt.Println("Failed to establish connection to Mongo server: ", err)
	}
	defer session.Close()
	if !bson.IsObjectIdHex(id){
		return "404" //NOT FOUND
	}
	oid := bson.ObjectIdHex(id)
	if err := session.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).RemoveId(oid); err != nil {
		log.Fatal(err)
		return "500" //INTERNAL SERVER ERROR
	}
	return "OK"
}


