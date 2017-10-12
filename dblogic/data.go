package dblogic

import (
	"gopkg.in/mgo.v2/bson"
)

//Struct (Data) to write and retrieve something from database
type Data struct {
	Id bson.ObjectId `bson:"_id,omitempty" json:"id"`
	Name string `bson:"name" json:"name"`
	Data_itself interface{} `bson:"data" json:"data_itself"`
}


//Config struct have connection parameters such as hosts and ports
type Config struct {
	MongoDBhost 	string 		`json:"MongoDBhost"`
	MongoDBname 	string 		`json:"MongoDBname"`
	MongoDBdocname 	string 		`json:"MongoDBdocname"`
	AerospikeHost 	string 		`json:"AerospikeHost"`
	AerospikePort 	string		`json:"AerospikePort"`
}