package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var config struct {
	DbLogin      string `json:"dbLogin"`
	DbPassword   string `json:"dbPassword"`
	DbHost       string `json:"dbHost"`
	DbDb         string `json:"dbDb"`
	DbPort       string `json:"dbPort"`
	LdapUser     string `json:"ldapUser"`
	LdapPassword string `json:"ldapPassword"`
	LdapServer   string `json:"ldapServer"`
	LdapBaseDN   string `json:"ldapBaseDN"`
	SessionKey   string `json:"sessionKey"`
	RecaptchaKey string `json:"recaptchaKey"`
}

type server struct {
	Db *sqlx.DB
}

type Pagination struct {
	CurrentPage int
	NextPage    int
	PrevPage    int
	LastPage    int
	Offset      int
	PerPage     int
}

var (
	configFile  = flag.String("config", "conf.json", "Where to read the config from")
	servicePort = flag.Int("port", 4001, "Service port number")
	store       = sessions.NewCookieStore([]byte(config.SessionKey))
)

func loadConfig() error {
	jsonData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, &config)
}

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
		log.Println("Creating directory for user files")
	}
}

func (s *server) paginationCalc(page, perPage int, table string) Pagination {
	var count int
	var pagination Pagination
	var err error
	if page < 1 {
		page = 1
	}
	pagination.CurrentPage = page
	pagination.PerPage = perPage
	pagination.Offset = perPage * (page - 1)
	switch table {
	case "wifiUsers":
		err = s.Db.Get(&count, "SELECT COUNT(*) as count FROM wifiUsers")
	case "memorandums":
		err = s.Db.Get(&count, "SELECT COUNT(*) as count FROM memorandums")
	}
	if err != nil {
		log.Println(err)
		return Pagination{}
	}
	if count > perPage*page {
		pagination.NextPage = pagination.CurrentPage + 1
		if pagination.NextPage != (count/perPage)+1 {
			pagination.LastPage = (count / perPage) + 1
		}
	}
	if pagination.CurrentPage > 1 {
		pagination.PrevPage = pagination.CurrentPage - 1
	}
	return pagination
}

func (s *server) saveAction(userName, ip, action, id, item string) error {
	_, err := s.Db.Query("INSERT INTO `actions` (username, ip, action, itemid, item) VALUES (?, ?, ?, ?, ?)", userName, ip, action, id, item)
	return err
}

func main() {
	checkFolders()
	flag.Parse()
	if err := loadConfig(); err != nil {
		log.Fatal(err)
	}
	log.Println("Config loaded from", *configFile)
	s := server{
		Db: sqlx.MustConnect("postgres", "host="+config.DbHost+" port="+config.DbPort+" user="+config.DbLogin+" dbname="+config.DbDb+" password="+config.DbPassword),
	}
	defer s.Db.Close()

	log.Printf("Connected to database on %s", config.DbHost)

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/userFiles/", userFilesHandler)
	http.HandleFunc("/generatePdf/", s.generatePdfHandler)
	http.HandleFunc("/generatedPdf/", s.generatedPdfHandler)

	http.HandleFunc("/admin/", s.adminHandler)
	http.HandleFunc("/admin/login/", loginHandler)
	http.HandleFunc("/admin/logout/", logoutHandler)
	http.HandleFunc("/admin/memorandums/", s.showMemorandumsHandler)
	http.HandleFunc("/admin/users/", s.usersHandler)
	http.HandleFunc("/admin/user/", s.userHandler)
	http.HandleFunc("/admin/checkMemorandum/", s.checkMemorandumHandler)
	http.HandleFunc("/admin/departments/", s.departmentsHandler)

	port := strconv.Itoa(*servicePort)
	log.Println("Server started at port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
