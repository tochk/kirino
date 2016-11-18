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

func indexHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded index page from %s", r.RemoteAddr)
	tmpl := template.New("Index page template")
	tmpl, err := template.ParseFiles("layouts/base.tmpl")
	if err != nil {
		log.Fatal(err)
	}
	temp := Temp{Content: "tochk"}
	err = tmpl.ExecuteTemplate(w, "base.tmpl", temp)
	if err != nil {
		log.Fatal(err)
	}
}

func generateLatexTable(list []UserData) Table {
	table := ""
	for _, tempData := range list {
		stringInTable := tempData.MacAddr + " & " + tempData.UserName + " & " + tempData.PhoneNumber + " & \\\\ \n \\hline \n"
		table += stringInTable
	}
	return Table{Table: table}
}

func generateLatexFile(list []UserData, hashStr string) {
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
	err = latexTemplate.ExecuteTemplate(outputLatexFile, "wifi.tex", generateLatexTable(list))
	if err != nil {
		log.Fatal(err)
	}
	outputLatexFile.Close()
	pathToTexFile := "userFiles\\" + hashStr + ".tex"
	cmd := exec.Command("pdflatex", "--interaction=errorstopmode", "--synctex=-1", "-output-directory=userFiles", pathToTexFile)
	cmd.Start()
}

func main() {
	log.Print("Starting...")
	err := loadConfig()
	if err != nil {
		log.Fatal(err)
	}
	generateLatexFile([]UserData{{MacAddr: "148814881488", UserName: "tochk imba", PhoneNumber: "88005553535"}}, "hashZHIRHY")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fs := http.FileServer(http.Dir("static"))
	http.Handle("/static/", http.StripPrefix("/static/", fs))

	http.ListenAndServe(":8080", nil)
}