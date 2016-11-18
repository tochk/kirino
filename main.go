package main

import (
	"log"
	"net/http"
	"flag"
	"io/ioutil"
	"encoding/json"
	"fmt"
)

var config struct {
	MysqlLogin    string `json:"mysqlLogin"`
	MysqlPassword string `json:"mysqlPassword"`
	MysqlHost     string `json:"mysqlHost"`
	MysqlDb       string `json:"mysqlDb"`
}

var (
	configFile = flag.String("config", "conf.json", "Where to read the config from")
)

func loadConfig() error {
	jsonData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, &config)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded index page from %s", r.RemoteAddr)
	fmt.Fprint(w, "test")
}

func main() {
	log.Print("Starting...")
	err := loadConfig()
	if err != nil {
		log.Fatal(err)
	} else {
		log.Print("Config loaded")
	}
	http.HandleFunc("/", indexHandler)
	log.Print("Starting server...")
	http.ListenAndServe(":8080", nil)
}