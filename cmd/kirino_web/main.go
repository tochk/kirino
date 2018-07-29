package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/departments"
	"github.com/tochk/kirino/generator"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/users"
)

var (
	configFile  = flag.String("config", "conf.json", "Where to read the config from")
	servicePort = flag.Int("port", 4001, "Service port number")
)

func filesHandler(w http.ResponseWriter, r *http.Request) {
	log.Println(r.URL.Path)
	path := "." + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	http.NotFound(w, r)
}

func checkFolders() {
	if file, err := os.Open("userFiles"); err != nil {
		file.Close()
		if err = os.Mkdir("userFiles", 0644); err != nil {
			log.Fatal(err)
		}
		log.Println("Creating directory for documents")
	}
}

func main() {
	log.Println("Checking folder for documents")
	checkFolders()
	flag.Parse()
	if err := server.LoadConfig(*configFile); err != nil {
		log.Fatal(err)
	}
	log.Println("Config loaded from", *configFile)

	server.ConnectToDb()
	defer server.Core.Db.Close()
	log.Printf("Connected to database on %s", server.Config.DbHost)

	router := mux.NewRouter().StrictSlash(true)

	router.Methods("GET").PathPrefix("/static/").HandlerFunc(filesHandler)
	router.Methods("GET").PathPrefix("/userFiles/").HandlerFunc(filesHandler)

	router.HandleFunc("/", memorandums.FormsHandler).Methods("GET")
	router.HandleFunc("/{type}/", memorandums.FormsHandler).Methods("GET")
	router.HandleFunc("/generate/{type}/", generator.GenerateHandler).Methods("POST")
	router.HandleFunc("/generated/{type}/{token}/", generator.GeneratedHandler).Methods("GET")

	router.HandleFunc("/admin/{type}/", auth.Handler)

	router.HandleFunc("/departments/view/{page:[0-9]+}", departments.Handler)

	router.HandleFunc("/wifi/memorandums/", memorandums.ListWifiHandler)
	router.HandleFunc("/wifi/memorandum/", memorandums.ViewWifiHandler)
	router.HandleFunc("/wifi/users/", users.WifiUsersHandler)
	router.HandleFunc("/wifi/user/", users.WifiUserHandler)

	router.HandleFunc("/ethernet/memorandums/", memorandums.ListEthernetHandler)
	router.HandleFunc("/ethernet/memorandum/", memorandums.ViewEthernetHandler)

	router.HandleFunc("/phone/memorandums/", memorandums.ListPhoneHandler)
	router.HandleFunc("/phone/memorandum/", memorandums.ViewPhoneHandler)

	router.HandleFunc("/domain/memorandums/", memorandums.ListDomainHandler)

	router.HandleFunc("/mail/memorandums/", memorandums.ListMailHandler)
	router.HandleFunc("/mail/memorandum/", memorandums.ViewMailHandler)

	port := strconv.Itoa(*servicePort)
	log.Println("Server started at port", port)
	if err := http.ListenAndServe(":"+port, router); err != nil {
		log.Fatal(err)
	}
}
