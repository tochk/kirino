package latex

import (
	"os"
	"os/exec"
	"strings"
	"text/template"

	log "github.com/sirupsen/logrus"
	"github.com/tochk/kirino/templates/html"
)

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

type MailMemorandum struct {
	Table        string
	MemorandumId int
	Target       string
	Department   string
}

type PhoneMemorandum struct {
	Table        string
	MemorandumId int
	Department   string
}

type EthernetMemorandum struct {
	Table        string
	MemorandumId int
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

func generateWifiLatexTable(list []html.WifiUser, memorandumId int) WifiMemorandum {
	table := ""
	for _, tempData := range list {
		stringInTable := tempData.MacAddress + " & " + tempData.UserName + " & " + tempData.PhoneNumber + " & \\\\ \n \\hline \n"
		table += stringInTable
	}
	return WifiMemorandum{Table: table, MemorandumId: memorandumId}
}

func generateLatexFileForWifiMemorandum(list []html.WifiUser, hashStr string, memorandumId int) (string, error) {
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

func GenerateWifiMemorandum(list []html.WifiUser, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForWifiMemorandum(list, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}

func GenerateDomainMemorandum(domain html.Domain, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForDomainMemorandum(domain, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}

func generateDomainLatexTable(domain html.Domain, memorandumId int) DomainMemorandum {
	table := ""
	stringInTable := domain.Name + " & " + domain.Hosting + " & " + domain.FIO + " & " + domain.Accounts + " \\\\ \n \\hline \n"
	table += stringInTable
	return DomainMemorandum{Table: table, MemorandumId: memorandumId, Target: domain.Target, Department: domain.Department}
}

func generateLatexFileForDomainMemorandum(domain html.Domain, hashStr string, memorandumId int) (string, error) {
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

func GenerateMailMemorandum(mail []html.Mail, info html.MailMemorandum, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForMailMemorandum(mail, info, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}

func generateMailLatexTable(mail []html.Mail, memorandumId int) MailMemorandum {
	table := ""
	for _, e := range mail {
		stringInTable := e.Mail + " & " + e.Name + " & " + e.Position + " \\\\ \n \\hline \n"
		table += stringInTable
	}
	return MailMemorandum{Table: table, MemorandumId: memorandumId}
}

func generateLatexFileForMailMemorandum(mail []html.Mail, info html.MailMemorandum, hashStr string, memorandumId int) (string, error) {
	memorandumTemplate, err := template.ParseFiles("templates/latex/mail.tex")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	mm := generateMailLatexTable(mail, memorandumId)
	mm.Target = info.Reason
	mm.Department = info.Department
	err = memorandumTemplate.Execute(outputLatexFile, mm)
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles/" + hashStr + ".tex"
	return pathToTexFile, nil
}

func GeneratePhoneMemorandum(phone []html.Phone, info html.PhoneMemorandum, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForPhoneMemorandum(phone, info, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}

func generatePhoneLatexTable(mail []html.Phone, memorandumId int) MailMemorandum {
	table := ""
	for _, e := range mail {
		access := "Не указано"
		switch e.Access {
		case 1:
			access = "Внутренний"
		case 2:
			access = "Городской"
		case 3:
			access = "Межгородской"
		case 4:
			access = "Международный"
		}
		stringInTable := e.Phone + " & " + access + " & " + e.Info + " \\\\ \n \\hline \n"
		table += stringInTable
	}
	return MailMemorandum{Table: table, MemorandumId: memorandumId}
}

func generateLatexFileForPhoneMemorandum(mail []html.Phone, info html.PhoneMemorandum, hashStr string, memorandumId int) (string, error) {
	memorandumTemplate, err := template.ParseFiles("templates/latex/phone.tex")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	mm := generatePhoneLatexTable(mail, memorandumId)
	mm.Department = info.Department
	err = memorandumTemplate.Execute(outputLatexFile, mm)
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles/" + hashStr + ".tex"
	return pathToTexFile, nil
}

func GenerateEthernetMemorandum(ethernet []html.Ethernet, info html.EthernetMemorandum, hashStr string, memorandumId int) error {
	path, err := generateLatexFileForEthernetMemorandum(ethernet, info, hashStr, memorandumId)
	if err != nil {
		return err
	}
	err = generatePdf(path)
	return err
}

func generateEthernetLatexTable(ethernet []html.Ethernet, memorandumId int) EthernetMemorandum {
	table := ""
	for _, e := range ethernet {
		stringInTable := e.Mac + " & " + e.Class + " & " + e.Building + " & " + e.Info + " \\\\ \n \\hline \n"
		table += stringInTable
	}
	return EthernetMemorandum{Table: table, MemorandumId: memorandumId}
}

func generateLatexFileForEthernetMemorandum(ethernet []html.Ethernet, info html.EthernetMemorandum, hashStr string, memorandumId int) (string, error) {
	memorandumTemplate, err := template.ParseFiles("templates/latex/ethernet.tex")
	if err != nil {
		return "", err
	}
	outputLatexFile, err := os.Create("userFiles/" + hashStr + ".tex")
	if err != nil {
		return "", err
	}
	defer outputLatexFile.Close()
	mm := generateEthernetLatexTable(ethernet, memorandumId)
	mm.Department = info.Department
	err = memorandumTemplate.Execute(outputLatexFile, mm)
	if err != nil {
		return "", err
	}
	pathToTexFile := "userFiles/" + hashStr + ".tex"
	return pathToTexFile, nil
}
