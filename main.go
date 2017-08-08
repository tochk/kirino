package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.stingr.net/stingray/kirino_wifi/latex"
	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"gopkg.in/ldap.v2"
	"database/sql"
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
}

type server struct {
	Db *sqlx.DB
}

type FullWifiUser struct {
	Id           int `db:"id"`
	MacAddress   string `db:"mac"`
	UserName     string `db:"username"`
	PhoneNumber  string `db:"phonenumber"`
	Hash         string `db:"hash"`
	MemorandumId int    `db:"memorandumid"`
	Accepted     int    `db:"accepted"`
	Disabled     int    `db:"disabled"`
	DepartmentId *int `db:"departmentid"`
}

type FullWifiMemorandum struct {
	Id           int  `db:"id"`
	AddTime      string `db:"addtime"`
	Accepted     int  `db:"accepted"`
	Disabled     int  `db:"disabled"`
	DepartmentId *int `db:"departmentid"`
}

type FullWifiMemorandumClientList struct {
	Clients     []FullWifiUser
	Memorandum  FullWifiMemorandum
	Departments []Department
}

type Department struct {
	Id       int64 `db:"id"`
	Name     string `db:"name"`
	Selected bool
}

type GeneratedPdfPage struct {
	Token      string
	Exist      []string
	ExistCount int
	Count      string
	IsAdmin    bool
}

type MemorandumsPage struct {
	Memorandums []FullWifiMemorandum
	Departments []Department
	Pagination  Pagination
}

type UsersPage struct {
	Users       []FullWifiUser
	Departments []Department
	Pagination  Pagination
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

func convertDataForDb(oldData latex.WifiUser, hash string, memorandumId int) FullWifiUser {
	return FullWifiUser{MacAddress: oldData.MacAddress,
		UserName: oldData.UserName,
		PhoneNumber: oldData.PhoneNumber,
		Hash: hash,
		MemorandumId: memorandumId,
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

func (s *server) tryWriteUserDataToDb(tx *sqlx.Tx, data []latex.WifiUser, hash string) (memorandumId int, err error) {
	if err = tx.Get(&memorandumId, "SELECT max(id) FROM memorandums"); err != nil {
		if err.Error() == "sql: Scan error on column index 0: converting driver.Value type <nil> (\"<nil>\") to a int: invalid syntax" {
			memorandumId = 0
		} else {
			return 0, err
		}
	}
	memorandumId++
	if _, err := tx.Exec(tx.Rebind("INSERT INTO memorandums (id, addtime) VALUES (?, current_date())"), memorandumId); err != nil {
		return 0, err
	}

	stmt, err := tx.PrepareNamed("INSERT INTO wifiUsers (mac, userName, phoneNumber, hash, memorandumId) VALUES (:mac, :username, :phonenumber, :hash, :memorandumid)")
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
	hashedStr := hex.EncodeToString(hasher.Sum(nil))
	if file, err := os.Open("userFiles\\" + hashStr + ".tex"); err == nil {
		file.Close()
		hashedStr = generateHash(hashedStr)
	}
	return hashedStr
}

func checkMacAddresses(list []latex.WifiUser) ([]latex.WifiUser, error) {
	newList := make([]latex.WifiUser, 0, len(list))
	regForMac, err := regexp.Compile("[^a-f0-9]+")
	regForName, err := regexp.Compile("[^а-яА-Яa-zA-Z \\-]+")
	regForPhone, err := regexp.Compile("[^0-9+\\-() ]+")
	if err != nil {
		log.Println(err)
		return newList, err
	}
	for _, user := range list {
		user.MacAddress = string(bytes.ToLower([]byte(user.MacAddress)))
		user.MacAddress = regForMac.ReplaceAllString(user.MacAddress, "")
		user.UserName = regForName.ReplaceAllString(user.UserName, "")
		user.PhoneNumber = regForPhone.ReplaceAllString(user.PhoneNumber, "")
		if len(user.MacAddress) != 12 {
			err = errors.New("Invalid mac-address")
			log.Println(err)
			return newList, err
		}
		newList = append(newList, user)
	}
	return newList, nil
}

func checkSingleMac(mac string) (string, error) {
	r, err := regexp.Compile("[^a-f0-9]+")
	if err != nil {
		return "", err
	}
	mac = string(bytes.ToLower([]byte(mac)))
	mac = r.ReplaceAllString(mac, "")
	if len(mac) != 12 {
		err = errors.New("Invalid mac-address")
		return "", err
	}
	return mac, nil
}

func checkSingleName(name string) (string, error) {
	regForName, err := regexp.Compile("[^а-яА-Яa-zA-ZёЁ \\-]+")
	if err != nil {
		return "", err
	}
	return regForName.ReplaceAllString(name, ""), nil
}

func checkSinglePhone(phone string) (string, error) {
	regForPhone, err := regexp.Compile("[^0-9+\\-() ]+")
	if err != nil {
		return "", err
	}
	return regForPhone.ReplaceAllString(phone, ""), nil
}

func (s *server) generatePdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}
	exist := make([]string, 0, 5)
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

	listForDb, err := checkMacAddresses(list)
	if err != nil {
		log.Println(err)
		return
	}
	listToWrite := make([]latex.WifiUser, 0, 5)
	for _, e := range listForDb {
		if _, err := s.getUserByMac(e.MacAddress); err == sql.ErrNoRows {
			listToWrite = append(listToWrite, e)
		} else {
			exist = append(exist, e.MacAddress)
		}
	}

	listToWrite2 := make([]latex.WifiUser, 0, 5)
	for i, e := range listToWrite {
		duplicate := false
		for i2, e2 := range listToWrite2 {
			if i != i2 && e.MacAddress == e2.MacAddress {
				duplicate = true
				break
			}
		}
		if !duplicate {
			listToWrite2 = append(listToWrite2, e)
		}
	}
	listToWrite = listToWrite2
	if len(listToWrite) > 0 {
		memorandumId, err := s.writeUserDataToDb(listToWrite, hash)
		if err != nil {
			log.Println(err)
			return
		}

		if err = latex.GenerateWifiMemorandum(listToWrite, hash, memorandumId); err != nil {
			log.Println(err)
			return
		}
	}
	if len(exist) == 0 {
		http.Redirect(w, r, "/generatedPdf/"+hash+"/0/"+strconv.Itoa(len(listToWrite)), 302)
	} else {
		http.Redirect(w, r, "/generatedPdf/"+hash+"/"+strings.Join(exist, ",")+"/"+strconv.Itoa(len(listToWrite)), 302)
	}
}

func (s *server) generatedPdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	memorandum := r.URL.Path[len("/generatedPdf/"):]
	splittedUrl := strings.Split(memorandum, "/")
	var page GeneratedPdfPage
	page.Token = splittedUrl[0]
	if splittedUrl[1] != "0" {
		page.Exist = strings.Split(splittedUrl[1], ",")
	}
	page.Count = splittedUrl[2]
	page.ExistCount = len(page.Exist)
	latexTemplate, err := template.ParseFiles("templates/html/generatedPdf.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}

	if session.Values["userName"] != nil {
		page.IsAdmin = true
	}

	if err = latexTemplate.Execute(w, page); err != nil {
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
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var pagination Pagination
	perPage := 50
	var memorandums []FullWifiMemorandum
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/memorandums/"):]
	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "save":
			if len(splittedUrl[1]) > 0 {
				if _, err := s.Db.Exec("UPDATE memorandums SET departmentid = $1 WHERE id = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
					log.Println(err)
					return
				}
				if _, err := s.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE memorandumid = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
					log.Println(err)
					return
				}
				http.Redirect(w, r, r.Referer(), 302)
				return
			}
		case "page":
			page, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			pagination = s.paginationCalc(page, perPage, "memorandums")
			memorandums, err = s.getMemorandums(pagination.PerPage, pagination.Offset)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
	latexTemplate, err := template.ParseFiles("templates/html/memorandums.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}

	if pagination.CurrentPage == 0 {
		memorandums, err = s.getMemorandums(50, 0)
		if err != nil {
			log.Println(err)
			return
		}
		pagination = s.paginationCalc(1, perPage, "memorandums")
	}

	departments, err := s.getDepartments()
	if err != nil {
		log.Println(err)
		return
	}

	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}

	if err = latexTemplate.Execute(w, MemorandumsPage{
		Memorandums: memorandums,
		Departments: departments,
		Pagination:  pagination,
	}); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) acceptMemorandum(id string) (err error) {
	if _, err = s.Db.Exec("UPDATE wifiUsers SET accepted = 1 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = s.Db.Exec("UPDATE memorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func (s *server) rejectMemorandum(id string) (err error) {
	if _, err = s.Db.Exec("UPDATE wifiUsers SET accepted = 2 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = s.Db.Exec("UPDATE memorandums SET accepted = 2 WHERE id = $1", id)
	return
}

func (s *server) checkMemorandumHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	memId := r.URL.Path[len("/admin/checkMemorandum/"):]
	if memId == "" {
		log.Println("Invalid memorandum id")
		return
	}

	if len(memId) > len("accept/") {
		if memId[0:len("accept/")] == "accept/" {
			memId = memId[len("accept/"):]
			if err := s.acceptMemorandum(memId); err != nil {
				log.Println(err)
				return
			}
		} else if memId[0:len("reject/")] == "reject/" {
			memId = memId[len("reject/"):]
			if err := s.rejectMemorandum(memId); err != nil {
				log.Println(err)
				return
			}
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	}
	clientsInMemorandum := make([]FullWifiUser, 0)
	if err := s.Db.Select(&clientsInMemorandum, "SELECT id, mac, userName, phoneNumber, hash, memorandumId, accepted, disabled, departmentid FROM wifiUsers WHERE memorandumId = $1", memId); err != nil {
		log.Println(err)
		return
	}

	departments, err := s.getDepartments()
	if err != nil {
		log.Println(err)
		return
	}

	var memorandum FullWifiMemorandum
	if err := s.Db.Get(&memorandum, "SELECT id, addTime, accepted, departmentid FROM memorandums WHERE id = $1", memId); err != nil {
		log.Println(err)
		return
	}

	latexTemplate, err := template.ParseFiles("templates/html/checkMemorandum.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err = latexTemplate.Execute(w, FullWifiMemorandumClientList{Clients: clientsInMemorandum, Memorandum: memorandum, Departments: departments}); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) adminHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] != nil {
		http.Redirect(w, r, "/admin/memorandums/", 302)
		return
	}
	latexTemplate, err := template.ParseFiles("templates/html/admin.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err = latexTemplate.Execute(w, nil); err != nil {
		log.Println(err)
		return
	}
}

func auth(login, password string) (string, error) {
	username := ""
	l, err := ldap.Dial("tcp", config.LdapServer)
	if err != nil {
		return username, err
	}
	defer l.Close()

	if l.Bind(config.LdapUser, config.LdapPassword); err != nil {
		return username, err
	}

	searchRequest := ldap.NewSearchRequest(
		config.LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(sAMAccountName="+login+"))",
		[]string{"cn"},
		nil,
	)

	if sr, err := l.Search(searchRequest); err != nil || len(sr.Entries) != 1 {
		err = errors.New("User not found")
		return username, err
	} else {
		username = sr.Entries[0].GetAttributeValue("cn")
	}

	err = l.Bind(username, password)

	return username, err
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	r.ParseForm()
	session, _ := store.Get(r, "applicationData")

	if userName, err := auth(r.Form["login"][0], r.Form["password"][0]); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin/", 302)
	} else {
		session, _ = store.Get(r, "applicationData")
		session.Values["userName"] = userName
		session.Save(r, w)
		http.Redirect(w, r, "/admin/memorandums/", 302)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	session.Values["userName"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/admin/", 302)
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

func (s *server) userHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}

	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/user/"):]
	var user FullWifiUser
	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "edit":
			if len(splittedUrl[1]) > 0 {
				userId, err := strconv.Atoi(splittedUrl[1])
				if err != nil {
					log.Println(err)
					return
				}
				user, err = s.getUser(userId)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case "save":
			clearMac, err := checkSingleMac(r.PostForm.Get("mac1"))
			clearName, err := checkSingleName(r.PostForm.Get("user1"))
			clearPhone, err := checkSinglePhone(r.PostForm.Get("tel1"))
			if err != nil {
				log.Println(err)
				return
			}
			_, err = s.Db.Exec("UPDATE wifiUsers SET mac = $1, username = $2, phonenumber = $3 WHERE id = $4", clearMac, clearName, clearPhone, splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			http.Redirect(w, r, "/admin/users/", 302)
			return
		}
	}

	latexTemplate, err := template.ParseFiles("templates/html/user.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err = latexTemplate.Execute(w, user); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) getUserList(limit, offset int) (userList []FullWifiUser, err error) {
	err = s.Db.Select(&userList, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid FROM wifiUsers ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func (s *server) getMemorandums(limit, offset int) (memorandums []FullWifiMemorandum, err error) {
	err = s.Db.Select(&memorandums, "SELECT id, addTime, accepted, departmentid FROM memorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func (s *server) getUser(id int) (user FullWifiUser, err error) {
	err = s.Db.Get(&user, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid FROM wifiUsers WHERE id = $1", id)
	return
}

func (s *server) getUserByMac(mac string) (user FullWifiUser, err error) {
	err = s.Db.Get(&user, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid FROM wifiUsers WHERE mac = $1", mac)
	return
}

func (s *server) setDisabled(status, id int) (err error) {
	_, err = s.Db.Exec("UPDATE wifiUsers SET disabled = $1 WHERE id = $2", status, id)
	return
}

func (s *server) setRejected(status, id int) (err error) {
	_, err = s.Db.Exec("UPDATE wifiUsers SET accepted = $1 WHERE id = $2", status, id)
	return
}

func (s *server) usersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/users/"):]
	var usersList []FullWifiUser
	var pagination Pagination
	perPage := 50
	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "savedept":
			if len(splittedUrl[1]) > 0 {
				if _, err := s.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE id = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
					log.Println(err)
					return
				}
				http.Redirect(w, r, r.Referer(), 302)
				return
			}
		case "accept":
			id, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			if err = s.setRejected(1, id); err != nil {
				log.Println(err)
				return
			}
			http.Redirect(w, r, r.Referer(), 302)
			return
		case "reject":
			id, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			if err = s.setRejected(2, id); err != nil {
				log.Println(err)
				return
			}
			http.Redirect(w, r, r.Referer(), 302)
			return
		case "enable":
			id, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			if err = s.setDisabled(0, id); err != nil {
				log.Println(err)
				return
			}
			http.Redirect(w, r, r.Referer(), 302)
			return
		case "disable":
			id, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			if err = s.setDisabled(1, id); err != nil {
				log.Println(err)
				return
			}
			http.Redirect(w, r, r.Referer(), 302)
			return
		case "page":
			page, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			pagination = s.paginationCalc(page, perPage, "wifiUsers")
			usersList, err = s.getUserList(pagination.PerPage, pagination.Offset)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
	latexTemplate, err := template.ParseFiles("templates/html/users.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}
	if pagination.CurrentPage == 0 {
		usersList, err = s.getUserList(50, 0)
		if err != nil {
			log.Println(err)
			return
		}
		pagination = s.paginationCalc(1, perPage, "wifiUsers")
	}

	for i, e := range usersList {
		if len(e.UserName) > 65 {
			usersList[i].UserName = e.UserName[:65] + "..."
		}
	}

	departments, err := s.getDepartments()
	if err != nil {
		log.Println(err)
		return
	}

	if err = latexTemplate.Execute(w, UsersPage{
		Users:       usersList,
		Departments: departments,
		Pagination:  pagination,
	}); err != nil {
		log.Println(err)
		return
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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		http.ServeFile(w, r, "/static/favicon.ico")
		return
	}
	log.Printf("Loaded %s page from %s", r.URL.Path, r.RemoteAddr)

	session, _ := store.Get(r, "applicationData")

	latexTemplate, err := template.ParseFiles("templates/html/index.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}

	isAdmin := false
	if session.Values["userName"] != nil {
		isAdmin = true
	}

	if err = latexTemplate.Execute(w, isAdmin); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) getDepartments() ([]Department, error) {
	var departments []Department
	err := s.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return departments, err
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

	port := strconv.Itoa(*servicePort)
	log.Println("Server started at port", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatal(err)
	}
}
