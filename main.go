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

type LatexData struct {
	Table        string
	MemorandumId int
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

func writeUserDataToDb(data []UserData, hash string) (int, error) {

	return -1, nil
}

func generatePdfHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded generatePdf page from %s", r.RemoteAddr)
	err := r.ParseForm()
	if err != nil {
		log.Println(err)
		return
	}
	var list []UserData
	for i := 1; i <= len(r.Form)/3; i++ {
		tempUserData := UserData{
			MacAddr:     r.PostFormValue("mac" + strconv.Itoa(i)),
			UserName:    r.PostFormValue("user" + strconv.Itoa(i)),
			PhoneNumber: r.PostFormValue("tel" + strconv.Itoa(i)),
		}
		list = append(list, tempUserData)
	}
	hasher := sha256.New()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	hashStr := r.PostFormValue("mac1") + strconv.Itoa(r1.Intn(1000000))
	hasher.Write([]byte(hashStr))
	hash := hex.EncodeToString(hasher.Sum(nil))
	memorandumId, err := writeUserDataToDb(list, hash)
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
	http.Redirect(w, r, "/userFiles/"+hash+".pdf", 302)
}

func generateLatexTable(list []UserData, memorandumId int) LatexData {
	table := ""
	for _, tempData := range list {
		stringInTable := tempData.MacAddr + " & " + tempData.UserName + " & " + tempData.PhoneNumber + " & \\\\ \n \\hline \n"
		table += stringInTable
	}
	return LatexData{Table: table, MemorandumId:memorandumId}
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
	http.HandleFunc("/userFiles/", userFilesHandler)
	http.HandleFunc("/generatePdf/", generatePdfHandler)
	log.Print("Server started at port 4001")
	err = http.ListenAndServe(":4001", nil)
	if err != nil {
		log.Fatal(err)
	}
}
