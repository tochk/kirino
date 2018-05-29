package generator

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
	"github.com/tochk/kirino/users"
)

type WifiUser = html.WifiUser

type GeneratedPdfPage struct {
	Token      string
	Exist      []string
	ExistCount int
	Count      string
	IsAdmin    bool
}

func writeUserDataToDb(data []latex.WifiUser, hash string) (int, error) {
	for {
		tx, err := server.Core.Db.Beginx()
		if err != nil {
			return 0, err
		}
		id, err := tryWriteUserDataToDb(tx, data, hash)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		if err = tx.Commit(); err == nil {
			return id, nil
		}
	}
}

func tryWriteUserDataToDb(tx *sqlx.Tx, data []latex.WifiUser, hash string) (memorandumId int, err error) {
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
		element.Hash = hash
		*element.MemorandumId = memorandumId
		if _, err = stmt.Exec(element); err != nil {
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

func checkWifiData(list []WifiUser) ([]WifiUser, error) {
	var err error
	for i, user := range list {
		user.MacAddress, err = check.Mac(user.MacAddress)
		if err != nil {
			return make([]WifiUser, 0), err
		}
		user.UserName = check.Name(user.UserName)
		user.PhoneNumber = check.Phone(user.PhoneNumber)
		list[i] = user
	}
	return list, nil
}

func WifiGenerateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}
	if err := memorandums.CheckRecaptcha(r.FormValue("g-recaptcha-response")); err != nil {
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

	list, err := checkWifiData(list)
	if err != nil {
		log.Println(err)
		return
	}

	for _, e := range list {
		if _, err := users.GetWifiUserByMac(e.MacAddress); err != sql.ErrNoRows {
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
		memorandumId, err := writeUserDataToDb(listToWrite, hash)
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

func WifiGeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/generatedPdf/"):]
	splittedUrl := strings.Split(memorandumInfo, "/")

	var exist []string
	if splittedUrl[1] != "0" {
		exist = strings.Split(splittedUrl[1], ",")
	}

	pageType := "wifi"
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	fmt.Fprint(w, html.GeneratedPage(pageType, splittedUrl[0], splittedUrl[2], exist))
}
