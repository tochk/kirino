package generator

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/common"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func generatePhone(form url.Values) (string, string, error) {
	info := html.PhoneMemorandum{
		Department: form.Get("num1"),
	}

	list := make([]html.Phone, 0, (len(form)-2)/3)
	for i := 1; i <= (len(form)-2)/4; i++ {
		access, err := strconv.Atoi(latex.TexEscape(form.Get("typePhone" + strconv.Itoa(i))))
		if err != nil {
			return "", "", err
		}
		tempUserData := html.Phone{
			Phone:  latex.TexEscape(form.Get("num" + strconv.Itoa(i))),
			Info:   latex.TexEscape(form.Get("room"+strconv.Itoa(i)) + " кабинет " + form.Get("build"+strconv.Itoa(i)) + " корпуса"),
			Access: access,
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(form.Get("postAdress1"))

	list, err := checkPhoneData(list)
	if err != nil {
		return "", "", err
	}

	memorandumId, err := writePhoneDataToDb(list, info, hash)
	if err != nil {
		return "", "", err
	}

	if err = latex.GeneratePhoneMemorandum(list, info, hash, memorandumId); err != nil {
		return "", "", err
	}

	return hash, "", nil
}

func checkPhoneData(list []html.Phone) ([]html.Phone, error) {
	for i, user := range list {
		user.Phone = check.All(user.Phone)
		user.Info = check.All(user.Info)
		list[i] = user
	}
	return list, nil
}

func writePhoneDataToDb(data []html.Phone, info html.PhoneMemorandum, hash string) (int, error) {
	var id int
	err := common.RunTx(context.Background(), server.Core.Db, func(tx *sqlx.Tx) error {
		var err error
		id, err = tryWritePhoneDataToDb(tx, data, info, hash)
		return err
	})
	if err != nil {
		return 0, err
	}
	return id, nil
}

func tryWritePhoneDataToDb(tx *sqlx.Tx, data []html.Phone, info html.PhoneMemorandum, hash string) (memorandumId int, err error) {
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
