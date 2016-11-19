package main

import (
	"log"
	"net/http"
	"flag"
	"io/ioutil"
	"encoding/json"
	"text/template"
	"os"
	"os/exec"
	"crypto/sha256"
	"encoding/hex"
	"math/rand"
	"time"
	"strconv"
	"fmt"
)

var config struct {
	MysqlLogin    string `json:"mysqlLogin"`
	MysqlPassword string `json:"mysqlPassword"`
	MysqlHost     string `json:"mysqlHost"`
	MysqlDb       string `json:"mysqlDb"`
}

type Temp struct {
	Content string
}

type Table struct {
	Table string
}

type UserData struct {
	MacAddr     string
	UserName    string
	PhoneNumber string
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

func generatePdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded generatePdf page from %s", r.RemoteAddr)
	err := r.ParseForm()
	if err != nil {
		log.Fatal(err)
	}
	var list []UserData
	for i := 1; i <= len(r.Form) / 3; i++ {
		tempData := UserData{MacAddr: r.PostFormValue("mac" + strconv.Itoa(i)), UserName: r.PostFormValue("user" + strconv.Itoa(i)), PhoneNumber: r.PostFormValue("tel" + strconv.Itoa(i))}
		list = append(list, tempData)
	}
	hasher := sha256.New()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	hashStr := r.PostFormValue("mac1") + strconv.Itoa(r1.Intn(1000000))
	log.Print(hashStr)
	hasher.Write([]byte(hashStr))
	hash := hex.EncodeToString(hasher.Sum(nil))
	log.Print(hash)
	pathToTex := generateLatexFile(list, hash)
	generatePdf(pathToTex)
	log.Print(list)
	fmt.Fprint(w, "<a href='/userFiles/" + hash + ".pdf'>Link</a>")
}

func generateLatexTable(list []UserData) Table {
	table := ""
	for _, tempData := range list {
		stringInTable := tempData.MacAddr + " & " + tempData.UserName + " & " + tempData.PhoneNumber + " & \\\\ \n \\hline \n"
		table += stringInTable
	}
	return Table{Table: table}
}

func generateLatexFile(list []UserData, hashStr string) string {
	log.Print("Generating latex file")
	latexTemplate := template.New("Latex template")
	latexTemplate, err := template.ParseFiles("latex/wifi.tex")
	if err != nil {
		log.Fatal(err)
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		log.Fatal(err)
	}
	defer outputLatexFile.Close()
	err = latexTemplate.ExecuteTemplate(outputLatexFile, "wifi.tex", generateLatexTable(list))
	if err != nil {
		log.Fatal(err)
	}

	pathToTexFile := "userFiles\\" + hashStr + ".tex"
	return pathToTexFile
}

func generatePdf(path string) {
	cmd := exec.Command("pdflatex", "--interaction=errorstopmode", "--synctex=-1", "-output-directory=userFiles", path)
	cmd.Start()
}

func userFilesHandler(w http.ResponseWriter, r *http.Request) {
	path := "." + r.URL.Path
	if f, err := os.Stat(path); err == nil && !f.IsDir() {
		http.ServeFile(w, r, path)
		return
	}
	http.NotFound(w, r)
}

func main() {
	log.Print("Starting...")
	err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))
	http.HandleFunc("/userFiles/", userFilesHandler)
	http.HandleFunc("/generatePdf/", generatePdfHandler)
	http.ListenAndServe(":8080", nil)
}