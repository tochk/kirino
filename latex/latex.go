package latex

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"
)

type WifiUser struct {
	MacAddress  string
	UserName    string
	PhoneNumber string
}

type WifiMemorandum struct {
	Table        string
	MemorandumId int
}

func TexEscape(s string) string {
	s = strings.Replace(s, "%", "\\%", -1)
	s = strings.Replace(s, "$", "\\$", -1)
	s = strings.Replace(s, "_", "\\_", -1)
	s = strings.Replace(s, "{", "\\{", -1)
	s = strings.Replace(s, "#", "\\#", -1)
	s = strings.Replace(s, "&", "\\&", -1)
	return s
}

func generateLatexTable(list []WifiUser, memorandumId int) WifiMemorandum {
	table := ""
	for _, tempData := range list {
		stringInTable := tempData.MacAddress + " & " + tempData.UserName + " & " + tempData.PhoneNumber + " & \\\\ \n \\hline \n"
		table += stringInTable
	}
	return WifiMemorandum{Table: table, MemorandumId: memorandumId}
}

func generateLatexFileForWifiMemorandum(list []WifiUser, hashStr string, memorandumId int) (string, error) {
	wifiMemorandumTemplate, err := template.ParseFiles("templates/latex/wifi.tex")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	err = wifiMemorandumTemplate.Execute(outputLatexFile, generateLatexTable(list, memorandumId))
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles/" + hashStr + ".tex"
	return pathToTexFile, nil
}

func generatePdf(path string) error {
	cmd := exec.Command("pdflatex", "--interaction=errorstopmode", "--synctex=-1", "-output-directory=userFiles", path)
	err := cmd.Run()
	if err != nil {
		log.Println("pdflatex", "--interaction=errorstopmode", "--synctex=-1", "-output-directory=userFiles", path)
	}
	return err
}

func GenerateWifiMemorandum(list []WifiUser, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForWifiMemorandum(list, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}
