package generator

import (
	"database/sql"
	"net/url"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
	"github.com/tochk/kirino/users"
)

func generateWifi(form url.Values) (string, string, error) {
	exist := make([]string, 0, 5)
	list := make([]latex.WifiUser, 0, 5)
	for i := 1; i <= len(form)/3; i++ {
		tempUserData := latex.WifiUser{
			MacAddress:  latex.TexEscape(form.Get("mac" + strconv.Itoa(i))),
			UserName:    latex.TexEscape(form.Get("user" + strconv.Itoa(i))),
			PhoneNumber: latex.TexEscape(form.Get("tel" + strconv.Itoa(i))),
		}
		list = append(list, tempUserData)
	}

	hash := generateHash(form.Get("mac1"))

	list, err := checkWifiData(list)
	if err != nil {
		return "", "", err
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
		memorandumId, err := writeWifiUserDataToDb(listToWrite, hash)
		if err != nil {
			return "", "", err
		}

		if err = latex.GenerateWifiMemorandum(listToWrite, hash, memorandumId); err != nil {
			return "", "", err
		}
	}

	existList := "0"
	if len(exist) != 0 {
		existList = strings.Join(exist, ",")
	}
	return hash, "?exist=" + existList + "&count=" + strconv.Itoa(len(listToWrite)), nil
}

func checkWifiData(list []html.WifiUser) ([]html.WifiUser, error) {
	var err error
	for i, user := range list {
		user.MacAddress, err = check.Mac(user.MacAddress)
		if err != nil {
			return nil, err
		}
		user.UserName = check.Name(user.UserName)
		user.PhoneNumber = check.Phone(user.PhoneNumber)
		list[i] = user
	}
	return list, nil
}

func writeWifiUserDataToDb(data []latex.WifiUser, hash string) (int, error) {
	for {
		tx, err := server.Core.Db.Beginx()
		if err != nil {
			return 0, err
		}
		id, err := tryWriteWifiUserDataToDb(tx, data, hash)
		if err != nil {
			tx.Rollback()
			return 0, err
		}
		if err = tx.Commit(); err == nil {
			return id, nil
		}
	}
}

func tryWriteWifiUserDataToDb(tx *sqlx.Tx, data []latex.WifiUser, hash string) (memorandumId int, err error) {
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
		element.MemorandumId = &memorandumId
		if _, err = stmt.Exec(element); err != nil {
			return 0, err
		}
	}

	return memorandumId, nil
}
