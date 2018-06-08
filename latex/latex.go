package latex

import (
	"log"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/tochk/kirino/templates/html"
)

type WifiUser = html.WifiUser
type Domain = html.Domain

type WifiMemorandum struct {
	Table        string
	MemorandumId int
}

type DomainMemorandum struct {
	Table        string
	MemorandumId int
	Target       string
	Department   string
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

func generateWifiLatexTable(list []WifiUser, memorandumId int) WifiMemorandum {
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
	err = wifiMemorandumTemplate.Execute(outputLatexFile, generateWifiLatexTable(list, memorandumId))
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

func GenerateDomainMemorandum(domain Domain, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForDomainMemorandum(domain, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}

func generateDomainLatexTable(domain Domain, memorandumId int) DomainMemorandum {
	table := ""
	stringInTable := domain.Name + " & " + domain.Hosting + " & " + domain.FIO + " & " + domain.Accounts + " \\\\ \n \\hline \n"
	table += stringInTable
	return DomainMemorandum{Table: table, MemorandumId: memorandumId, Target: domain.Target, Department: domain.Department}
}

func generateLatexFileForDomainMemorandum(domain Domain, hashStr string, memorandumId int) (string, error) {
	memorandumTemplate, err := template.ParseFiles("templates/latex/domain.tex")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	err = memorandumTemplate.Execute(outputLatexFile, generateDomainLatexTable(domain, memorandumId))
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles/" + hashStr + ".tex"
	return pathToTexFile, nil
}
