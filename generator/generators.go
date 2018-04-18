package generator

import (
	"bytes"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"git.stingr.net/stingray/kirino_wifi/latex"
	"github.com/tochk/kirino_wifi/templates/html"
	"github.com/jmoiron/sqlx"
)

type GeneratedPdfPage struct {
	Token      string
	Exist      []string
	ExistCount int
	Count      string
	IsAdmin    bool
}

func convertDataForDb(oldData latex.WifiUser, hash string, memorandumId int) FullWifiUser {
	return FullWifiUser{MacAddress: oldData.MacAddress,
		UserName: oldData.UserName,
		PhoneNumber: oldData.PhoneNumber,
		Hash: hash,
		MemorandumId: &memorandumId,
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
	regForName, err := regexp.Compile("[^а-яА-Яa-zA-Z \\.\\-]+")
	regForPhone, err := regexp.Compile("[^0-9+\\-() ]+")
	if err != nil {
		log.Println(err)
		return newList, err
	}
	for _, user := range list {
		user.MacAddress = string(bytes.ToLower([]byte(user.MacAddress)))
		user.MacAddress = regForMac.ReplaceAllString(user.MacAddress, "")
		user.UserName = strings.Replace(user.UserName, "Ё", "Е", -1)
		user.UserName = strings.Replace(user.UserName, "ё", "е", -1)
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
	regForName, err := regexp.Compile("[^а-яА-Яa-zA-Z \\.\\-]+")
	if err != nil {
		return "", err
	}
	name = strings.Replace(name, "Ё", "Е", -1)
	name = strings.Replace(name, "ё", "е", -1)
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
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}
	if err := checkRecaptcha(r.FormValue("g-recaptcha-response")); err != nil {
		log.Println(err)
		fmt.Fprint(w, "Капча введена неправильно")
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

	list, err := checkMacAddresses(list)
	if err != nil {
		log.Println(err)
		return
	}

	for _, e := range list {
		if _, err := s.getUserByMac(e.MacAddress); err != sql.ErrNoRows {
			exist = append(exist, e.MacAddress)
		}
	}

	listToWrite := make([]latex.WifiUser, 0, 5)
	for i, e := range list {
		duplicate := false
		for i2, e2 := range listToWrite {
			if i != i2 && e.MacAddress == e2.MacAddress {
				duplicate = true
				break
			}
		}
		if !duplicate {
			listToWrite = append(listToWrite, e)
		}
	}

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
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/generatedPdf/"):]
	splittedUrl := strings.Split(memorandumInfo, "/")

	var exist []string
	if splittedUrl[1] != "0" {
		exist = strings.Split(splittedUrl[1], ",")
	}

	fmt.Fprint(w, html.GeneratedPage("Доступ к WiFi сети СГУ", isAdmin(r), splittedUrl[0], splittedUrl[2], exist))
}
