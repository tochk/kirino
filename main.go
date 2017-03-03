package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"text/template"
	"github.com/jmoiron/sqlx"
	htemplate "html/template"
	"time"
	_ "github.com/lib/pq"
)

var config struct {
	DbLogin    string `json:"dbLogin"`
	DbPassword string `json:"dbPassword"`
	DbHost     string `json:"dbHost"`
	DbDb       string `json:"dbDb"`
	DbPort     string `json:"dbPort"`
}

type Temp struct {
	Content string
}

type LatexData struct {
	Table        string
	MemorandumId int
}

type UserData struct {
	MacAddr     string
	UserName    string
	PhoneNumber string
}

type DataForDb struct {
	MacAddr      string `db:"mac"`
	UserName     string `db:"userName"`
	PhoneNumber  string `db:"phoneNumber"`
	Hash         string `db:"hash"`
	MemorandumId int    `db:"memorandumId"`
}

type MemorandumData struct {
	UserCount int `db:"userCount"`
}

type MemorandumDataForPage struct {
	Id        int `db:"id"`
	UserCount *int `db:"userCount"`
}

type DataForWriteToAdminTempltate struct {
	Memorandums []MemorandumDataForPage
}

type DataForWriteToCheckMemTempltate struct {
	Clients []DataForDb
}

type DataForGeneratedPdf struct {
	PdfUrl string
}

var (
	configFile = flag.String("config", "conf.json", "Where to read the config from")
)

func texEscape(s string) string {
	s = strings.Replace(s, "%", "\\%", -1)
	s = strings.Replace(s, "$", "\\$", -1)
	s = strings.Replace(s, "_", "\\_", -1)
	s = strings.Replace(s, "{", "\\{", -1)
	s = strings.Replace(s, "#", "\\#", -1)
	s = strings.Replace(s, "&", "\\&", -1)
	return s
}

func loadConfig() error {
	jsonData, err := ioutil.ReadFile(*configFile)
	if err != nil {
		return err
	}
	return json.Unmarshal(jsonData, &config)
}

func convertDataForDb(oldData UserData, hash string, memorandumId int) DataForDb {
	return DataForDb{MacAddr: oldData.MacAddr,
		UserName:         oldData.UserName,
		PhoneNumber:      oldData.PhoneNumber,
		Hash:             hash,
		MemorandumId:     memorandumId,
	}
}

func (s *server) writeUserDataToDb(data []UserData, hash string) (int, error) {
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

func (s *server) tryWriteUserDataToDb(tx *sqlx.Tx, data []UserData, hash string) (int, error) {
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

func (s *server) generatePdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded generatePdf page from %s", r.RemoteAddr)
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	var list []UserData
	for i := 1; i <= len(r.Form)/3; i++ {
		tempUserData := UserData{
			MacAddr:     texEscape(r.PostFormValue("mac" + strconv.Itoa(i))),
			UserName:    texEscape(r.PostFormValue("user" + strconv.Itoa(i))),
			PhoneNumber: texEscape(r.PostFormValue("tel" + strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}
	hasher := sha256.New()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	hashStr := r.PostFormValue("mac1") + strconv.Itoa(r1.Intn(1000000))
	hasher.Write([]byte(hashStr))
	hash := hex.EncodeToString(hasher.Sum(nil))
	memorandumId, err := s.writeUserDataToDb(list, hash)
	if err != nil {
		log.Println(err)
		return
	}
	pathToTex, err := generateLatexFile(list, hash, memorandumId)
	if err != nil {
		log.Println(err)
		return
	}
	err = generatePdf(pathToTex)
	if err != nil {
		log.Println(err)
		return
	}
	w.Header().Set("Location", "/userFiles/"+hash+".pdf")
	_, err = template.ParseFiles("userFiles/" + hash + ".pdf")
	if err != nil {
		log.Println(err)
		return
	}
	http.Redirect(w, r, "/generatedPdf/?hash="+hash, 302)
}

func (s *server) generatedPdfHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	latexTemplate := htemplate.New("Main design")
	latexTemplate, err := htemplate.ParseFiles("templates/generatedPdf.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	err = latexTemplate.Execute(w, DataForGeneratedPdf{PdfUrl:r.Form.Get("hash")})
	if err != nil {
		log.Println(err)
		return
	}
}

func generateLatexTable(list []UserData, memorandumId int) LatexData {
	table := ""
	for _, tempData := range list {
		stringInTable := tempData.MacAddr + " & " + tempData.UserName + " & " + tempData.PhoneNumber + " & \\\\ \n \\hline \n"
		table += stringInTable
	}
	return LatexData{Table: table, MemorandumId: memorandumId}
}

func generateLatexFile(list []UserData, hashStr string, memorandumId int) (string, error) {
	latexTemplate := template.New("Latex template")
	latexTemplate, err := template.ParseFiles("latex/wifi.tex")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	err = latexTemplate.ExecuteTemplate(outputLatexFile, "wifi.tex", generateLatexTable(list, memorandumId))
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles\\" + hashStr + ".tex"
	return pathToTexFile, nil
}

func generatePdf(path string) error {
	cmd := exec.Command("pdflatex", "--interaction=errorstopmode", "--synctex=-1", "-output-directory=userFiles", path)
	err := cmd.Run()
	return err
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
	latexTemplate := htemplate.New("Main design")
	latexTemplate, err := htemplate.ParseFiles("templates/design.tmpl")
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
	if err := tx.Select(&memorandums, "SELECT id, userCount FROM memorandums"); err != nil {
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

func (s *server) checkMemorandumHandler(w http.ResponseWriter, r *http.Request) {
	memId := r.URL.Path[len("/checkMemorandum/"):]
	tx, err := s.Db.Beginx()
	if err != nil {
		log.Println(err)
		return
	}
	clientsInMemorandum := make([]DataForDb, 0)
	if err := tx.Select(&clientsInMemorandum, "SELECT mac, userName, phoneNumber, hash, memorandumId FROM wifiUsers WHERE memorandumId = $1", memId); err != nil {
		log.Println(err)
		return
	}
	latexTemplate, err := htemplate.ParseFiles("templates/checkMem.tmpl")
	if err != nil {
		log.Println(err)
		return
	}
	err = latexTemplate.Execute(w, DataForWriteToCheckMemTempltate{Clients: clientsInMemorandum})
	if err != nil {
		log.Println(err)
		return
	}
	tx.Commit()
}

type server struct {
	Db *sqlx.DB
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
	http.HandleFunc("/memorandums/", s.showMemorandumsHandler)
	http.HandleFunc("/checkMemorandum/", s.checkMemorandumHandler)
	http.HandleFunc("/generatedPdf/", s.generatedPdfHandler)
	log.Print("Server started at port 4001")
	err = http.ListenAndServe(":4001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
