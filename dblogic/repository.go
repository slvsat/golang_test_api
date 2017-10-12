package dblogic

import (
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/aerospike/aerospike-client-go"
	"encoding/json"
	"strconv"
	"errors"
)

//All manipulations (Create, Read, Update, Delete) will be realized through this struct
//to have always opened connection to the database we have pointers to session, client, etc.
type Repository struct{
	mongoSession *mgo.Session
	client *aerospike.Client
	policy *aerospike.WritePolicy
}


//Default config for connections (Specified for MacOS - Vagrant Aerospike and MongoDB)
//you can just change config file to have another
var Conf Config = Config {
	"localhost:27017",
	"Test",
	"testlist",
	"172.28.128.3",
	"3000",
}


//Initializer
func NewRepository() *Repository {
	session, err := mgo.Dial(Conf.MongoDBhost)
	if err != nil {
		log.Println("Cannot make session to MongoDB! ", err)
		panic(err)
	}
	port, _ := strconv.Atoi(Conf.AerospikePort)
	conClient, err := aerospike.NewClient(Conf.AerospikeHost, port)
	if err != nil{
		log.Println("Cannot connect to Aerospike! ", err)
		panic(err)
	}
	conPolicy := aerospike.NewWritePolicy(0, 10)
	return &Repository{
		mongoSession: session,
		client: conClient,
		policy: conPolicy,
	}
}

//Clearing MongoDB table (if have any)
func (r *Repository) ClearTable() {
	r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).RemoveAll(nil)
}

//Setting config if have a config file
func (r *Repository) SetConfig(config Config){
	Conf = config
}

//Using aerospike as a cache, so key is the current url and data returned from MongoDB
func (r *Repository) writeDataToAerospike(key *aerospike.Key, data []Data) error {
	dataToWrite, _ := json.Marshal(data)
	bins := aerospike.BinMap{
		"bin1" : string(dataToWrite),
	}
	err := r.client.Put(r.policy, key, bins)
	if err != nil {
		return err
	}
	return nil
}

//Getting data from Aerospike by a specified key
func (r *Repository) getFromAerospike(key *aerospike.Key) (string, error) {
	rec, err := r.client.Get(nil, key)
	if err != nil {
		log.Println("Error while getting data from aerospike", err)
		return "", err
	}
	return rec.Bins["bin1"].(string), nil
}

//Parsing query string, if have any
func parseQuery(q string) (bson.M, error){
	if q != "" {
		outQuery := bson.M{}
		err := bson.UnmarshalJSON([]byte(q), &outQuery)
		if err != nil{
			log.Println("Error while UnmarshalingJSON ", err)
			return nil, err
		}
		//log.Println(outQuery)
		return outQuery, nil
	}
	return nil, nil
}

//Taking url and id as strings, returning data from Aerospike if exist, otherwise from MongoDB with specified id
func (r *Repository) GetDataById(url string, id string) ([]byte, error) {
	key, err := aerospike.NewKey("test", "aerospike", url)
	if err != nil {
		log.Println("Error while creating a key (get data by id) ", err)
	}
	exist, err := r.client.Exists(nil, key)
	if err != nil {
		log.Println("Given key doesn't exist! ", err)
	}
	result := make([]Data, 0)
	if exist == false{
		c := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname)
		if err := c.FindId(bson.ObjectIdHex(id)).All(&result); err != nil {
			log.Println("Cannot find item by ID ", err)
		}
		err := r.writeDataToAerospike(key, result)
		if err != nil {
			log.Println("function writeDataToAerospike crashed ", err)
			return nil, err
		}
	}else {
		out, err := r.getFromAerospike(key)
		if err != nil {
			log.Println("Cannot get data from Aerospike ", err)
		}
		return []byte(out), err
	}
	return json.Marshal(result)
}

//Returning whole data from Aerospike if exist, otherwise from MongoDB directly
func (r *Repository) GetData(url string, query string) ([]byte, error) {
	key, err := aerospike.NewKey("test", "aerospike", url)
	if err != nil {
		log.Println("Error while creating a key (aerospike)", err)
		return nil, err
	}

	exist, err := r.client.Exists(nil, key)
	if err != nil {
		log.Println("Key doesn't exist! ", err)
	}
	results := make([]Data, 0)
	if exist == false {
		c := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname)
		parsedQuery, err := parseQuery(query)
		if err != nil {
			log.Println("Error while parsing query! ", err)
			return nil, err
		}
		if err := c.Find(parsedQuery).All(&results); err != nil {
			log.Println("Failed to write results:", err)
		}
		r.writeDataToAerospike(key, results)
	}else {
		out, err := r.getFromAerospike(key)
		if err != nil {
			log.Println("Cannot get data from Aerospike! ", err)
			return nil, err
		}
		return []byte(out), nil
	}
	output, err := json.Marshal(results)
	if err != nil {
		log.Println("Error while Marshaling data", err)
		return nil, err
	}
	return output, nil
}


//Adding data to MongoDB
func (r *Repository) AddData(data Data) (string, error){
	data.Id = bson.NewObjectId()
	err := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).Insert(data)
	if err != nil {
		return "", err
	}
	return data.Id.Hex(), nil
}

//Updating data in MongoDB with specified id
func (r *Repository) UpdateData(data Data) error {
	err := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).Update(bson.M{"_id": data.Id}, bson.M{ "$set": bson.M{"name": data.Name, "data": data.Data_itself }})
	if err != nil {
		log.Println("Cannot update item ", err)
		return err
	}
	return nil
}

//Deleting data by id
//If ID is not ObjectIdHex returning an error
func (r *Repository) DeleteData(id string) (string, error){
	if !bson.IsObjectIdHex(id) {
		return "404", errors.New("ID is not ObjectIdHex! ")
	}
	oid := bson.ObjectIdHex(id)
	if err := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).RemoveId(oid); err != nil {
		log.Println("Error while removing item! ", err)
		return "500", err //INTERNAL SERVER ERROR
	}
	return "OK", nil
}