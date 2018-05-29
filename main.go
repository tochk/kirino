package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/generator"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/server"
)

var (
	configFile  = flag.String("config", "conf.json", "Where to read the config from")
	servicePort = flag.Int("port", 4001, "Service port number")
)

func userFilesHandler(w http.ResponseWriter, r *http.Request) {
	path := "." + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	http.NotFound(w, r)
}

func checkFolders() {
	if file, err := os.Open("documents"); err != nil {
		file.Close()
		if err = os.Mkdir("documents", 0644); err != nil {
			log.Fatal(err)
		}
		log.Println("Creating directory for documents")
	}
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	//todo some index page
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

	http.HandleFunc("/", indexHandler)
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/userFiles/", userFilesHandler)

	http.HandleFunc("/", memorandums.WifiHandler)

	http.HandleFunc("/wifi/generate/", generator.WifiGenerateHandler)
	http.HandleFunc("/wifi/generated/", generator.WifiGeneratedHandler)

	http.HandleFunc("/admin/", auth.Handler)
	http.HandleFunc("/admin/memorandums/", showMemorandumsHandler)
	http.HandleFunc("/admin/users/", usersHandler)
	http.HandleFunc("/admin/user/", userHandler)
	http.HandleFunc("/admin/checkMemorandum/", checkMemorandumHandler)
	http.HandleFunc("/admin/departments/", departmentsHandler)

	port := strconv.Itoa(*servicePort)
	log.Println("Server started at port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
