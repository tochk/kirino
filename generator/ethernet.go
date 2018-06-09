package generator

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type EthernetMemorandum = html.EthernetMemorandum

func EthernetGenerateHandler(w http.ResponseWriter, r *http.Request) {
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

	info := EthernetMemorandum{
		Department: r.PostForm.Get("dep1"),
	}

	list := make([]Ethernet, 0, (len(r.Form)-2)/3)
	for i := 1; i <= (len(r.Form)-2)/4; i++ {
		tempUserData := Ethernet{
			Mac :  latex.TexEscape(r.PostFormValue("mac" + strconv.Itoa(i))),
			Info:   latex.TexEscape(r.PostFormValue("descrip"+strconv.Itoa(i))),
			Class: latex.TexEscape(r.PostForm.Get("room" + strconv.Itoa(i))),
			Building: latex.TexEscape(r.PostForm.Get("build"+ strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(r.PostFormValue("mac1"))

	list, err := checkEthernetData(list)
	if err != nil {
		log.Println(err)
		return
	}

	memorandumId, err := writeEthernetDataToDb(list, info, hash)
	if err != nil {
		log.Println(err)
		return
	}

	if err = latex.GenerateEthernetMemorandum(list, info, hash, memorandumId); err != nil {
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/ethernet/generated/"+hash+"/", 302)

}

func checkEthernetData(list []Ethernet) ([]Ethernet, error) {
	var err error
	for i, user := range list {
		user.Mac, err = check.Mac(user.Mac)
		if err != nil {
			return nil, err
		}
		user.Info = check.All(user.Info)
		user.Class = check.All(user.Class)
		user.Building = check.All(user.Building)
		list[i] = user
	}
	return list, nil
}

func writeEthernetDataToDb(data []Ethernet, info EthernetMemorandum, hash string) (int, error) {
	tx, err := server.Core.Db.Beginx()
	if err != nil {
		return 0, err
	}
	id, err := tryWriteEthernetDataToDb(tx, data, info, hash)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err == nil {
		return id, nil
	}
	return 0, nil
}

func tryWriteEthernetDataToDb(tx *sqlx.Tx, data []Ethernet, info EthernetMemorandum, hash string) (memorandumId int, err error) {
	if err = tx.Get(&memorandumId, "SELECT max(id) FROM ethmemorandums"); err != nil {
		if err.Error() == "sql: Scan error on column index 0: converting driver.Value type <nil> (\"<nil>\") to a int: invalid syntax" {
			memorandumId = 0
		} else {
			return 0, err
		}
	}
	memorandumId++
	if _, err := tx.Exec("INSERT INTO ethmemorandums (id, addtime, department, hash) VALUES ($1, current_date(), $2, $3)", memorandumId, info.Department, hash); err != nil {
		return 0, err
	}

	for _, element := range data {
		_, err := tx.Exec("INSERT INTO ethusers (mac, class, building, info, memorandumId) VALUES ($1, $2, $3, $4, $5)", element.Mac, element.Class, element.Building, element.Info, memorandumId)
		if err != nil {
			return 0, err
		}
	}

	return memorandumId, nil
}

func EthernetGeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/ethernet/generated/"):]
	splittedUrl := strings.Split(memorandumInfo, "/")

	pageType := "ethernet"
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	fmt.Fprint(w, html.EthernetGeneratedPage(pageType, splittedUrl[0]))
}
