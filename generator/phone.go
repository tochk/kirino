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

type PhoneMemorandum = html.PhoneMemorandum

func PhoneGenerateHandler(w http.ResponseWriter, r *http.Request) {
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

	info := PhoneMemorandum{
		Department: r.PostForm.Get("num1"),
	}

	list := make([]Phone, 0, (len(r.Form)-2)/3)
	for i := 1; i <= (len(r.Form)-2)/4; i++ {
		access, err := strconv.Atoi(latex.TexEscape(r.PostFormValue("typePhone" + strconv.Itoa(i))))
		if err != nil {
			log.Println(err)
			return
		}
		tempUserData := Phone{
			Phone:  latex.TexEscape(r.PostFormValue("num" + strconv.Itoa(i))),
			Info:   latex.TexEscape(r.PostFormValue("room"+strconv.Itoa(i)) + " кабинет " + r.PostForm.Get("build"+strconv.Itoa(i)) + " корпуса"),
			Access: access,
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(r.PostFormValue("postAdress1"))

	list, err := checkPhoneData(list)
	if err != nil {
		log.Println(err)
		return
	}

	memorandumId, err := writePhoneDataToDb(list, info, hash)
	if err != nil {
		log.Println(err)
		return
	}

	if err = latex.GeneratePhoneMemorandum(list, info, hash, memorandumId); err != nil {
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/phone/generated/"+hash+"/", 302)

}

func checkPhoneData(list []Phone) ([]Phone, error) {
	for i, user := range list {
		user.Phone = check.All(user.Phone)
		user.Info = check.All(user.Info)
		list[i] = user
	}
	return list, nil
}

func writePhoneDataToDb(data []Phone, info PhoneMemorandum, hash string) (int, error) {
	tx, err := server.Core.Db.Beginx()
	if err != nil {
		return 0, err
	}
	id, err := tryWritePhoneDataToDb(tx, data, info, hash)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err == nil {
		return id, nil
	}
	return 0, nil
}

func tryWritePhoneDataToDb(tx *sqlx.Tx, data []Phone, info PhoneMemorandum, hash string) (memorandumId int, err error) {
	if err = tx.Get(&memorandumId, "SELECT max(id) FROM phonememorandums"); err != nil {
		if err.Error() == "sql: Scan error on column index 0: converting driver.Value type <nil> (\"<nil>\") to a int: invalid syntax" {
			memorandumId = 0
		} else {
			return 0, err
		}
	}
	memorandumId++
	if _, err := tx.Exec("INSERT INTO phonememorandums (id, addtime, department, exist, hash) VALUES ($1, current_date(), $2, 0, $3)", memorandumId, info.Department, hash); err != nil {
		return 0, err
	}

	for _, element := range data {
		_, err := tx.Exec("INSERT INTO phoneusers (phone, info, access, memorandumId) VALUES ($1, $2, $3, $4)", element.Phone, element.Info, element.Access, memorandumId)
		if err != nil {
			return 0, err
		}
	}

	return memorandumId, nil
}

func PhoneGeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/phone/generated/"):]
	splittedUrl := strings.Split(memorandumInfo, "/")

	pageType := "phone"
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	fmt.Fprint(w, html.PhoneGeneratedPage(pageType, splittedUrl[0]))
}
