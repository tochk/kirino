package memorandums

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/departments"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type memAccepted struct {
	MemorandumId int `db:"memorandumid"`
	MaxAccepted  int `db:"maxaccepted"`
	MinAccepted  int `db:"minaccepted"`
}

func WifiMemorandumsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	vars := mux.Vars(r)

	switch vars["action"] {
	case "save":
		r.ParseForm()
		err := saveWifiMemorandumDepartment(vars["num"], r.PostForm.Get("department"))
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
	case "view":
		memorandums, departmentList, paging, err := viewWifiMemorandums(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.WifiMemorandumsPage(memorandums, departmentList, paging))
	case "show":
		memorandum, clientsInMemorandum, departmentList, err := showWifiMemorandum(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.WifiMemorandumPage(memorandum, clientsInMemorandum, departmentList))
	case "accept":
		if err := acceptWifiMemorandum(vars["num"]); err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
	case "reject":
		if err := rejectWifiMemorandum(vars["num"]); err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
	}
}

func showWifiMemorandum(id string) (memorandum html.Memorandum, clientsInMemorandum []html.WifiUser, departmentList []html.Department, err error) {
	clientsInMemorandum, err = getWifiMemorandumClients(id)
	if err != nil {
		return
	}

	departmentList, err = departments.GetAll()
	if err != nil {
		return
	}

	memorandum, err = getWifiMemorandum(id)
	return
}

func viewWifiMemorandums(pageNum string) (memorandums []html.Memorandum, departmentList []html.Department, paging html.Pagination, err error) {
	page, err := strconv.Atoi(pageNum)
	if err != nil {
		return
	}

	count, err := getWifiMemorandumsCount()
	if err != nil {
		return
	}

	paging = pagination.Calc(page, count)

	memorandums, err = getWifiMemorandums(paging.PerPage, paging.Offset)
	if err != nil {
		return
	}

	departmentList, err = departments.GetAll()
	if err != nil {
		return
	}

	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}

	return
}

func saveWifiMemorandumDepartment(id, department string) (err error) {
	if _, err = server.Core.Db.Exec("UPDATE memorandums SET departmentid = $1 WHERE id = $2", department, id); err != nil {
		return
	}
	_, err = server.Core.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE memorandumid = $2", department, id)
	return
}

func acceptWifiMemorandum(id string) (err error) {
	if _, err = server.Core.Db.Exec("UPDATE wifiUsers SET accepted = 1 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = server.Core.Db.Exec("UPDATE memorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func rejectWifiMemorandum(id string) (err error) {
	if _, err = server.Core.Db.Exec("UPDATE wifiUsers SET accepted = 2 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = server.Core.Db.Exec("UPDATE memorandums SET accepted = 2 WHERE id = $1", id)
	return
}

func getWifiMemorandumsCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM memorandums")
	return
}

func getWifiMemorandums(limit, offset int) (memorandums []html.Memorandum, err error) {
	if err = server.Core.Db.Select(&memorandums, "SELECT id, addTime, accepted, departmentid FROM memorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset); err != nil {
		return
	}
	var accepted []memAccepted
	err = server.Core.Db.Select(&accepted, "SELECT max(accepted) as maxAccepted, min(accepted) as minAccepted, memorandumid FROM wifiusers WHERE memorandumid IN (SELECT id FROM memorandums) GROUP BY memorandumid ORDER BY memorandumid desc LIMIT $1 OFFSET $2 ", limit, offset)
	for i, e := range memorandums {
		for _, em := range accepted {
			if e.Id == em.MemorandumId && em.MaxAccepted != em.MinAccepted {
				memorandums[i].Accepted = 3
			}
		}
	}
	return
}

func getWifiMemorandum(id string) (memorandum html.Memorandum, err error) {
	err = server.Core.Db.Get(&memorandum, "SELECT id, addTime, accepted, departmentid FROM memorandums WHERE id = $1", id)
	return
}

func getWifiMemorandumClients(id string) (clientsInMemorandum []html.WifiUser, err error) {
	err = server.Core.Db.Select(&clientsInMemorandum, "SELECT id, mac, userName, phoneNumber, hash, memorandumId, accepted, disabled, departmentid FROM wifiUsers WHERE memorandumId = $1", id)
	return
}

func CheckWifiMemorandumAccepted(userId int) (err error) {
	_, err = server.Core.Db.Exec("UPDATE memorandums SET accepted = 1 "+
		"WHERE id = (SELECT memorandumid FROM wifiusers WHERE id = $1) AND "+
		"(SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1))"+
		" - (SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1) AND accepted = 1) = 0;", userId)
	if err != nil {
		return
	}
	_, err = server.Core.Db.Exec("UPDATE memorandums SET accepted = 2 "+
		"WHERE id = (SELECT memorandumid FROM wifiusers WHERE id = $1) AND "+
		"(SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1))"+
		" - (SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1) AND accepted = 2) = 0;", userId)
	return
}
