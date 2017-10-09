package dblogic

import (
	"log"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"github.com/aerospike/aerospike-client-go"
	"encoding/json"
	"strconv"
)
type Repository struct{
	mongoSession *mgo.Session
	client *aerospike.Client
	policy *aerospike.WritePolicy
}

var Conf Config = Config {
	"localhost:27017",
	"Test",
	"testlist",
	"172.28.128.3",
	"3000",
}

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

func (r *Repository) ClearTable() {
	r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).RemoveAll(nil)
}

func (r *Repository) SetConfig(config Config){
	Conf = config
}

func (r *Repository) writeDataToAerospike(key *aerospike.Key, data []Data) bool {
	dataToWrite, _ := json.Marshal(data)
	bins := aerospike.BinMap{
		"bin1" : string(dataToWrite),
	}
	err := r.client.Put(r.policy, key, bins)
	if err != nil {
		return false
	}
	return true
}

func (r *Repository) getFromAerospike(key *aerospike.Key) (string, error) {
	rec, err := r.client.Get(nil, key)
	if err != nil {
		log.Println("Error while getting data from aerospike", err)
	}
	return rec.Bins["bin1"].(string), err
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
		r.writeDataToAerospike(key, result)
	}else {
		out, err := r.getFromAerospike(key)
		if err != nil {
			log.Println("Cannot get data from Aerospike ", err)
		}
		return []byte(out), err
	}
	output, err := json.Marshal(result)
	return output, err
}

func (r *Repository) GetData(url string, query string) ([]byte, error) {
	key, err := aerospike.NewKey("test", "aerospike", url)
	if err != nil {
		log.Println("Error while creating a key (aerospike)", err)
		panic(err)
	}

	exist, err := r.client.Exists(nil, key)
	if err != nil {
		log.Println("Key doesn't exist! ", err)
	}
	results := make([]Data, 0)
	if exist == false {
		c := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname)
		if err := c.Find(parseQuery(query)).All(&results); err != nil {
			log.Println("Failed to write results:", err)
		}
		r.writeDataToAerospike(key, results)
	}else {
		out, err := r.getFromAerospike(key)
		if err != nil {
			log.Println("Cannot get data from Aerospike! ", err)
		}
		return []byte(out), err
	}
	output, err := json.Marshal(results)
	return output, err
}


func (r *Repository) AddData(data Data) (string, error){
	data.Id = bson.NewObjectId()
	err := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).Insert(data)
	return data.Id.Hex(), err
}

func (r *Repository) UpdateData(data Data) (bool, error) {
	err := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).Update(bson.M{"_id": data.Id}, bson.M{ "$set": bson.M{"name": data.Name, "data": data.Data_itself }})
	if err != nil {
		log.Println("Cannot update item ", err)
		return false, err
	}
	return true, err
}

func (r *Repository) DeleteData(id string) string{
	if !bson.IsObjectIdHex(id) {
		return "404" //NOT FOUND
	}
	oid := bson.ObjectIdHex(id)
	if err := r.mongoSession.DB(Conf.MongoDBname).C(Conf.MongoDBdocname).RemoveId(oid); err != nil {
		log.Println("Error while removing item! ", err)
		return "500" //INTERNAL SERVER ERROR
	}
	return "OK"
}