package main

import (
	"go.api/server"
	"github.com/urfave/cli"
	"fmt"
	"os"
	"io/ioutil"
	"encoding/json"
	"go.api/dblogic"
	"errors"
)

var conf dblogic.Config
var (
	PATH string = ""
	VERSION string = "0.1"
)

var flags []cli.Flag = []cli.Flag {
	cli.StringFlag{
		Name:        "config, c",
		Usage:       "Load configuration from `FILE`",
		Destination: &PATH,
	},
	cli.StringFlag{
		Name: "mghost",
		Usage: "MongoDB Hostname to connect",
		Destination: &conf.MongoDBhost,
	},
	cli.StringFlag{
		Name: "mgname",
		Usage: "MongoDB Name to connect",
		Destination: &conf.MongoDBname,
	},
	cli.StringFlag{
		Name: "mgdocname",
		Usage: "MongoDB DocName to connect",
		Destination: &conf.MongoDBdocname,
	},
	cli.StringFlag{
		Name:        "arsphost",
		Usage:       "Aerospike Hostname to connect",
		Destination: &conf.AerospikeHost,
	},
	cli.StringFlag{
		Name: "arspport",
		Usage: "Aerospike Port number to connect",
		Destination: &conf.AerospikePort,
	},
}


func main(){
	app := cli.NewApp()
	app.Name = "test RESTful api"
	app.Usage = "simple CRUD - mangoDB and aerospike as a cache"
	app.Version = VERSION
	app.Flags = flags
	app.Action = runAPI

	fmt.Println(app.Run(os.Args))
}

func ExtractConfig(path string, config *dblogic.Config) dblogic.Config{
	file, _ := ioutil.ReadFile(path)
	json.Unmarshal(file, &config)
	return *config
}

func runAPI(*cli.Context) error {

	if PATH != "" {
		ExtractConfig(PATH, &conf)
	}
	if conf.MongoDBhost == "" {
		return errors.New("Nothing found in MongoDB Hostname variable")
	}
	if conf.AerospikeHost == "" {
		return errors.New("Nothing found in Aerospike Hostname variable")
	}

	server.Run(conf)

	return nil
}