package main

import (
	"flag"
	"log"
	"net/http"
	"os"
	"strconv"

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

func userFilesHandler(w http.ResponseWriter, r *http.Request) {
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

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/userFiles/", userFilesHandler)

	http.HandleFunc("/", memorandums.WifiHandler)

	http.HandleFunc("/wifi/generate/", generator.WifiGenerateHandler)
	http.HandleFunc("/wifi/generated/", generator.WifiGeneratedHandler)

	http.HandleFunc("/admin/", auth.Handler)
	http.HandleFunc("/admin/departments/", departments.Handler)

	http.HandleFunc("/admin/wifi/memorandums/", memorandums.ListWifiHandler)
	http.HandleFunc("/admin/wifi/memorandum/", memorandums.ViewWifiHandler)
	http.HandleFunc("/admin/wifi/users/", users.WifiUsersHandler)
	http.HandleFunc("/admin/wifi/user/", users.WifiUserHandler)

	http.HandleFunc("/admin/ethernet/memorandums/", memorandums.ListEthernetHandler)
	http.HandleFunc("/admin/ethernet/memorandum/", memorandums.ViewEthernetHandler)

	http.HandleFunc("/admin/phone/memorandums/", memorandums.ListPhoneHandler)
	http.HandleFunc("/admin/phone/memorandum/", memorandums.ViewPhoneHandler)

	http.HandleFunc("/admin/domain/memorandums/", memorandums.ListDomainHandler)

	http.HandleFunc("/admin/mail/memorandums/", memorandums.ListMailHandler)
	http.HandleFunc("/admin/mail/memorandum/", memorandums.ViewMailHandler)

	http.HandleFunc("/ethernet/", memorandums.EthernetHandler)
	http.HandleFunc("/ethernet/generate/", generator.EthernetGenerateHandler)
	http.HandleFunc("/ethernet/generated/", generator.EthernetGeneratedHandler)

	http.HandleFunc("/phone/", memorandums.PhoneHandler)
	http.HandleFunc("/phone/generate/", generator.PhoneGenerateHandler)
	http.HandleFunc("/phone/generated/", generator.PhoneGeneratedHandler)

	http.HandleFunc("/domain/", memorandums.DomainHandler)
	http.HandleFunc("/domain/generate/", generator.DomainGenerateHandler)
	http.HandleFunc("/domain/generated/", generator.DomainGeneratedHandler)

	http.HandleFunc("/mail/", memorandums.MailHandler)
	http.HandleFunc("/mail/generate/", generator.MailGenerateHandler)
	http.HandleFunc("/mail/generated/", generator.MailGeneratedHandler)

	port := strconv.Itoa(*servicePort)
	log.Println("Server started at port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
