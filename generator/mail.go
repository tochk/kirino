package generator

import (
	"net/url"
	"strconv"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func generateMail(form url.Values) (string, string, error) {
	info := html.MailMemorandum{
		Reason:     form.Get("target_mail"),
		Department: form.Get("dep1"),
	}

	list := make([]html.Mail, 0, (len(form)-3)/3)
	for i := 1; i <= (len(form)-3)/3; i++ {
		tempUserData := html.Mail{
			Mail:     latex.TexEscape(form.Get("postAdress" + strconv.Itoa(i))),
			Name:     latex.TexEscape(form.Get("postName" + strconv.Itoa(i))),
			Position: latex.TexEscape(form.Get("postPosition" + strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(form.Get("postAdress1"))

	list, err := checkMailData(list)
	if err != nil {
		return "", "", err
	}

	memorandumId, err := writeMailDataToDb(list, info, hash)
	if err != nil {
		return "", "", err
	}

	if err = latex.GenerateMailMemorandum(list, info, hash, memorandumId); err != nil {
		return "", "", err
	}

	return hash, "", nil
}

func checkMailData(list []html.Mail) ([]html.Mail, error) {
	for i, user := range list {
		user.Mail = check.All(user.Mail)
		user.Position = check.All(user.Position)
		user.Name = check.All(user.Name)
		list[i] = user
	}
	return list, nil
}

func writeMailDataToDb(data []html.Mail, info html.MailMemorandum, hash string) (int, error) {
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

func tryWriteMailDataToDb(tx *sqlx.Tx, data []html.Mail, info html.MailMemorandum, hash string) (memorandumId int, err error) {
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
