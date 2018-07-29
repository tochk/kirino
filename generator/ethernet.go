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

func generateEthernet(form url.Values) (string, string, error) {
	info := html.EthernetMemorandum{
		Department: form.Get("dep1"),
	}

	list := make([]html.Ethernet, 0, (len(form)-2)/3)
	for i := 1; i <= (len(form)-2)/4; i++ {
		tempUserData := html.Ethernet{
			Mac:      latex.TexEscape(form.Get("mac" + strconv.Itoa(i))),
			Info:     latex.TexEscape(form.Get("descrip" + strconv.Itoa(i))),
			Class:    latex.TexEscape(form.Get("room" + strconv.Itoa(i))),
			Building: latex.TexEscape(form.Get("build" + strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(form.Get("mac1"))

	list, err := checkEthernetData(list)
	if err != nil {
		return "", "", err
	}

	memorandumId, err := writeEthernetDataToDb(list, info, hash)
	if err != nil {
		return "", "", err
	}

	if err = latex.GenerateEthernetMemorandum(list, info, hash, memorandumId); err != nil {
		return "", "", err
	}

	return hash, "", nil
}

func checkEthernetData(list []html.Ethernet) ([]html.Ethernet, error) {
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

func writeEthernetDataToDb(data []html.Ethernet, info html.EthernetMemorandum, hash string) (int, error) {
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

func tryWriteEthernetDataToDb(tx *sqlx.Tx, data []html.Ethernet, info html.EthernetMemorandum, hash string) (memorandumId int, err error) {
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
