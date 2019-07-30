package main

import (
	"flag"
	"net/http"
	"os"

	"github.com/GeertJohan/go.rice"
	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/departments"
	"github.com/tochk/kirino/generator"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/users"
)

var (
	configFile  = flag.String("config", "conf.json", "Where to read the config from")
	servicePort = flag.String("port", ":4001", "Service port number")
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

	static := rice.MustFindBox("../../static")
	s := http.StripPrefix("/static/", http.FileServer(static.HTTPBox()))
	router.PathPrefix("/static/").Handler(s)

	router.Methods("GET").PathPrefix("/userFiles/").HandlerFunc(filesHandler)

	router.HandleFunc("/", memorandums.FormsHandler).Methods("GET")
	router.HandleFunc("/{type}/", memorandums.FormsHandler).Methods("GET")
	router.HandleFunc("/generate/{type}/", generator.GenerateHandler).Methods("POST")
	router.HandleFunc("/generated/{type}/{token}/", generator.GeneratedHandler).Methods("GET")

	router.HandleFunc("/admin/{type}/", auth.Handler)

	router.HandleFunc("/wifi/memorandums/{action}/{num:[0-9]+}", memorandums.WifiMemorandumsHandler)
	router.HandleFunc("/wifi/users/{action}/{num:[0-9]+}", users.WifiUsersHandler)

	router.HandleFunc("/departments/{action}/{num:[0-9]+}", departments.Handler)

	router.HandleFunc("/ethernet/memorandums/{action}/{num:[0-9]+}", memorandums.ListEthernetHandler)

	router.HandleFunc("/phone/memorandums/{action}/{num:[0-9]+}", memorandums.ListPhoneHandler)

	router.HandleFunc("/domain/memorandums/{action}/{num:[0-9]+}", memorandums.ListDomainHandler)

	router.HandleFunc("/mail/memorandums/{action}/{num:[0-9]+}", memorandums.ListMailHandler)

	log.Println("Server started at", *servicePort)
	if err := http.ListenAndServe(*servicePort, router); err != nil {
		log.Fatal(err)
	}
}
