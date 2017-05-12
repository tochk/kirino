package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	htemplate "html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"text/template"
	"time"

	"git.stingr.net/stingray/kirino_wifi/latex"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/ldap.v2"
)

var config struct {
	DbLogin      string `json:"dbLogin"`
	DbPassword   string `json:"dbPassword"`
	DbHost       string `json:"dbHost"`
	DbDb         string `json:"dbDb"`
	DbPort       string `json:"dbPort"`
	LdapUser     string `json:"ldapUser"`
	LdapPassword string `json:"ldapPassword"`
}

type Temp struct {
	Content string
}

type DataForDb struct {
	MacAddr      string `db:"mac"`
	UserName     string `db:"userName"`
	PhoneNumber  string `db:"phoneNumber"`
	Hash         string `db:"hash"`
	MemorandumId int    `db:"memorandumId"`
	Accepted     int    `db:"accepted"`
	Disabled     int    `db:"disabled"`
}

type MemorandumData struct {
	UserCount int `db:"userCount"`
}

type MemorandumDataForPage struct {
	Id        int  `db:"id"`
	UserCount *int `db:"userCount"`
	Accepted  int  `db:"accepted"`
	Disabled  int  `db:"disabled"`
}

type DataForWriteToAdminTempltate struct {
	Memorandums []MemorandumDataForPage
}

type DataForWriteToCheckMemTempltate struct {
	Clients []DataForDb
	Id      string
}

type DataForGeneratedPdf struct {
	PdfUrl string
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

func convertDataForDb(oldData latex.WifiUser, hash string, memorandumId int) DataForDb {
	return DataForDb{MacAddr: oldData.MacAddress,
		UserName:         oldData.UserName,
		PhoneNumber:      oldData.PhoneNumber,
		Hash:             hash,
		MemorandumId:     memorandumId,
	}
}

func (s *server) writeUserDataToDb(data []latex.WifiUser, hash string) (int, error) {
	for {
		tx, err := s.Db.Beginx()
		if err != nil {
			return 0, err
		}
		id, err := s.tryWriteUserDataToDb(tx, data, hash)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		if err = tx.Commit(); err == nil {
			return id, nil
		}
	}
}

func (s *server) tryWriteUserDataToDb(tx *sqlx.Tx, data []latex.WifiUser, hash string) (int, error) {
	var memorandumId int
	if err := tx.Get(&memorandumId, "SELECT max(id) FROM memorandums"); err != nil {
		return 0, err
	}
	memorandumId++
	userCount := len(data)
	if _, err := tx.Exec(tx.Rebind("INSERT INTO memorandums (id, UserCount) VALUES (?, ?)"), memorandumId, userCount); err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareNamed("INSERT INTO wifiUsers (mac, userName, phoneNumber, hash, memorandumId) VALUES (:mac, :userName, :phoneNumber, :hash, :memorandumId)")
	if err != nil {
		return 0, err
	}
	defer stmt.Close()
	for _, element := range data {
		if _, err = stmt.Exec(convertDataForDb(element, hash, memorandumId)); err != nil {
			return 0, err
		}
	}

	return memorandumId, nil
}

func generateHash(firstMac string) string {
	hasher := sha256.New()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	hashStr := firstMac + strconv.Itoa(r1.Intn(1000000))
	hasher.Write([]byte(hashStr))
	return hex.EncodeToString(hasher.Sum(nil))
}

func (s *server) generatePdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}
	list := make([]latex.WifiUser, 0, 5)
	for i := 1; i <= len(r.Form)/3; i++ {
		tempUserData := latex.WifiUser{
			MacAddress:  latex.TexEscape(r.PostFormValue("mac" + strconv.Itoa(i))),
			UserName:    latex.TexEscape(r.PostFormValue("user" + strconv.Itoa(i))),
			PhoneNumber: latex.TexEscape(r.PostFormValue("tel" + strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(r.PostFormValue("mac1"))

	memorandumId, err := s.writeUserDataToDb(list, hash)
	if err != nil {
		log.Println(err)
		return
	}

	if err = latex.GenerateWifiMemorandum(list, hash, memorandumId); err != nil {
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/generatedPdf/"+hash, 302)
}

func (s *server) generatedPdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	token := r.URL.Path[len("/generatedPdf/"):]
	latexTemplate, err := htemplate.ParseFiles("templates/generatedPdf.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	err = latexTemplate.Execute(w, DataForGeneratedPdf{PdfUrl: token})
	if err != nil {
		log.Println(err)
		return
	}
}

func userFilesHandler(w http.ResponseWriter, r *http.Request) {
	path := "." + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	http.NotFound(w, r)
}

func (s *server) showMemorandumsHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	latexTemplate, err := htemplate.ParseFiles("templates/memorandums.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	tx, err := s.Db.Beginx()
	if err != nil {
		log.Println(err)
		return
	}
	memorandums := make([]MemorandumDataForPage, 0)
	if err := tx.Select(&memorandums, "SELECT id, userCount, accepted FROM memorandums ORDER BY id DESC"); err != nil {
		log.Println(err)
		return
	}
	err = latexTemplate.Execute(w, DataForWriteToAdminTempltate{Memorandums: memorandums})
	if err != nil {
		log.Println(err)
		return
	}
	tx.Commit()
}

func (s *server) addMemorandum(id string) (err error) {
	_, err = s.Db.Exec("UPDATE wifiUsers SET accepted = 1 WHERE memorandumId = $1", id)
	if err != nil {
		return
	}
	_, err = s.Db.Exec("UPDATE memorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func (s *server) checkMemorandumHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	memId := r.URL.Path[len("/admin/checkMemorandum/"):]
	if len(memId) > len("add/") {
		if memId[0:len("add/")] == "add/" {
			memId = memId[len("add/"):]
			err := s.addMemorandum(memId)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
	tx, err := s.Db.Beginx()
	if err != nil {
		log.Println(err)
		return
	}
	clientsInMemorandum := make([]DataForDb, 0)
	if err := tx.Select(&clientsInMemorandum, "SELECT mac, userName, phoneNumber, hash, memorandumId, accepted, disabled FROM wifiUsers WHERE memorandumId = $1", memId); err != nil {
		log.Println(err)
		return
	}
	latexTemplate, err := htemplate.ParseFiles("templates/checkMem.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	err = latexTemplate.Execute(w, DataForWriteToCheckMemTempltate{Clients: clientsInMemorandum, Id: memId})
	if err != nil {
		log.Println(err)
		return
	}
	tx.Commit()
}

type server struct {
	Db *sqlx.DB
}

func (s *server) adminHandler(w http.ResponseWriter, r *http.Request) {
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] != nil {
		http.Redirect(w, r, "/admin/memorandums/", 302)
		return
	}
	log.Println("Loaded admin login page from " + r.RemoteAddr)
	latexTemplate, err := template.ParseFiles("templates/admin.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	err = latexTemplate.Execute(w, nil)
	if err != nil {
		log.Println(err)
		return
	}
}

func auth(login, password string) (username string, err error) {
	l, err := ldap.Dial("tcp", fmt.Sprintf("%s:%d", "main.sgu.ru", 389))
	if err != nil {
		return
	}
	defer l.Close()

	err = l.Bind(config.LdapUser, config.LdapPassword)
	if err != nil {
		return
	}

	searchRequest := ldap.NewSearchRequest(
		"dc=main,dc=sgu,dc=ru",
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(sAMAccountName="+login+"))",
		[]string{"cn"},
		nil,
	)

	sr, err := l.Search(searchRequest)
	if err != nil {
		return
	}

	if len(sr.Entries) == 1 {
		username = sr.Entries[0].GetAttributeValue("cn")
	} else {
		err = errors.New("User not found")
		return
	}

	err = l.Bind(username, password)
	if err != nil {
		return
	}

	return
}

var store = sessions.NewCookieStore([]byte("applicationDataLP"))

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Loaded login page from " + r.RemoteAddr)
	r.ParseForm()
	session, _ := store.Get(r, "applicationData")
	userName, err := auth(r.Form["login"][0], r.Form["password"][0])
	if err != nil {
		http.Redirect(w, r, "/admin/", 302)
	} else {
		session, _ = store.Get(r, "applicationData")
		session.Values["userName"] = userName
		session.Save(r, w)
		http.Redirect(w, r, "/admin/memorandums/", 302)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("Loaded logout page from " + r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	session, _ = store.Get(r, "applicationData")
	session.Values["userName"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/admin/", 302)
}

func main() {
	err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	s := server{
		Db: sqlx.MustConnect("postgres", "host="+config.DbHost+" port="+config.DbPort+" user="+config.DbLogin+" dbname="+config.DbDb+" password="+config.DbPassword),
	}
	defer s.Db.Close()

	http.HandleFunc("/userFiles/", userFilesHandler)
	http.HandleFunc("/generatePdf/", s.generatePdfHandler)
	http.HandleFunc("/generatedPdf/", s.generatedPdfHandler)

	http.HandleFunc("/admin/", s.adminHandler)
	http.HandleFunc("/admin/login/", loginHandler)
	http.HandleFunc("/admin/logout/", logoutHandler)
	http.HandleFunc("/admin/memorandums/", s.showMemorandumsHandler)
	http.HandleFunc("/admin/checkMemorandum/", s.checkMemorandumHandler)

	log.Print("Server started at port 4001")
	err = http.ListenAndServe(":4001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
