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

type MailMemorandum = html.MailMemorandum

func MailGenerateHandler(w http.ResponseWriter, r *http.Request) {
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

	info := MailMemorandum{
		Reason:     r.PostForm.Get("target_mail"),
		Department: r.PostForm.Get("dep1"),
	}

	list := make([]Mail, 0, (len(r.Form)-3)/3)
	for i := 1; i <= (len(r.Form)-3)/3; i++ {
		tempUserData := Mail{
			Mail:     latex.TexEscape(r.PostFormValue("postAdress" + strconv.Itoa(i))),
			Name:     latex.TexEscape(r.PostFormValue("postName" + strconv.Itoa(i))),
			Position: latex.TexEscape(r.PostFormValue("postPosition" + strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(r.PostFormValue("postAdress1"))

	list, err := checkMailData(list)
	if err != nil {
		log.Println(err)
		return
	}

	memorandumId, err := writeMailDataToDb(list, info, hash)
	if err != nil {
		log.Println(err)
		return
	}

	if err = latex.GenerateMailMemorandum(list, info, hash, memorandumId); err != nil {
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/mail/generated/"+hash+"/", 302)

}

func checkMailData(list []Mail) ([]Mail, error) {
	for i, user := range list {
		user.Mail = check.All(user.Mail)
		user.Position = check.All(user.Position)
		user.Name = check.All(user.Name)
		list[i] = user
	}
	return list, nil
}

func writeMailDataToDb(data []Mail, info MailMemorandum, hash string) (int, error) {
	tx, err := server.Core.Db.Beginx()
	if err != nil {
		return 0, err
	}
	id, err := tryWriteMailDataToDb(tx, data, info, hash)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err == nil {
		return id, nil
	}
	return 0, nil
}

func tryWriteMailDataToDb(tx *sqlx.Tx, data []Mail, info MailMemorandum, hash string) (memorandumId int, err error) {
	if err = tx.Get(&memorandumId, "SELECT max(id) FROM mailmemorandums"); err != nil {
		if err.Error() == "sql: Scan error on column index 0: converting driver.Value type <nil> (\"<nil>\") to a int: invalid syntax" {
			memorandumId = 0
		} else {
			return 0, err
		}
	}
	memorandumId++
	if _, err := tx.Exec("INSERT INTO mailmemorandums (id, addtime, department, reason, hash) VALUES ($1, current_date(), $2, $3, $4)", memorandumId, info.Department, info.Reason, hash); err != nil {
		return 0, err
	}

	for _, element := range data {
		_, err := tx.Exec("INSERT INTO mailusers (mail, name, position, memorandumId) VALUES ($1, $2, $3, $4)", element.Mail, element.Name, element.Position, memorandumId)
		if err != nil {
			return 0, err
		}
	}

	return memorandumId, nil
}

func MailGeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/mail/generated/"):]
	splittedUrl := strings.Split(memorandumInfo, "/")

	pageType := "mail"
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	fmt.Fprint(w, html.MailGeneratedPage(pageType, splittedUrl[0]))
}
