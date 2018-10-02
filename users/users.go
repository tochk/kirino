package users

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/departments"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

const (
	acceptUser  = 1
	rejectUser  = 2
	enableUser  = 0
	disableUser = 1
)

func WifiUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	vars := mux.Vars(r)

	switch vars["action"] {
	case "save_dept":
		err := saveDepartment(vars["num"], r.PostForm.Get("department"))
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "accept":
		err := acceptOrRejectWifiUser(vars["num"], acceptUser)
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "reject":
		err := acceptOrRejectWifiUser(vars["num"], rejectUser)
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "enable":
		err := enableOrDisableWifiUser(vars["num"], enableUser)
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "disable":
		err := enableOrDisableWifiUser(vars["num"], disableUser)
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "view":
		usersList, departmentsList, paging, err := viewWifiUsers(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.WifiUsersPage(usersList, departmentsList, paging))
	case "search":
		usersList, departmentsList, err := getSearchResult(r.URL.Query())
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.WifiUsersPage(usersList, departmentsList, html.Pagination{CurrentPage:1}))
	case "edit":
		if r.Method == "GET" {
			user, depts, err := getUser(vars["num"])
			if err != nil {
				fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
				log.Print(err)
				return
			}
			fmt.Fprint(w, html.WifiUserPage(user, depts))
			return
		} else {
			err := updateUser(vars["num"], r.PostForm.Get("mac1"), r.PostForm.Get("user1"), r.PostForm.Get("tel1"))
			if err != nil {
				fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
				log.Print(err)
				return
			}
			http.Redirect(w, r, "/admin/wifi/users/", 302)
			return
		}
	}
}

func getUserList(limit, offset int) (userList []html.WifiUser, err error) {
	err = server.Core.Db.Select(&userList, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getUserCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM wifiUsers")
	return
}

func GetWifiUserById(id int) (user html.WifiUser, err error) {
	err = server.Core.Db.Get(&user, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers WHERE id = $1", id)
	return
}

func GetWifiUserByMac(mac string) (user html.WifiUser, err error) {
	err = server.Core.Db.Get(&user, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers WHERE accepted = 1 AND mac = $1", mac)
	return
}

func setDisabled(status, id int) (err error) {
	_, err = server.Core.Db.Exec("UPDATE wifiUsers SET disabled = $1 WHERE id = $2", status, id)
	return
}

func setAccepted(status, id int) (err error) {
	_, err = server.Core.Db.Exec("UPDATE wifiUsers SET accepted = $1 WHERE id = $2", status, id)
	return
}

func getSearchResult(values url.Values) (userList []html.WifiUser, departmentsList []html.Department, err error) {
	err = server.Core.Db.Select(&userList, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers WHERE mac LIKE CONCAT(CONCAT('%', $1), '%') AND username LIKE CONCAT(CONCAT('%', $2), '%') AND phonenumber LIKE CONCAT(CONCAT('%', $3), '%') ORDER BY id DESC ", values.Get("mac"), values.Get("name"), values.Get("phone"))
	if err != nil {
		return nil, nil, err
	}
	departmentsList, err = departments.GetAll()
	if err != nil {
		return nil, nil, err
	}
	return
}

func filterUserNames(users []html.WifiUser) {
	for i, e := range users {
		if len(e.UserName) > 50 {
			users[i].UserName = string([]rune(e.UserName)[:50]) + "..."
		}
	}
}

func acceptOrRejectWifiUser(idString string, num int) error {
	id, err := strconv.Atoi(idString)
	if err != nil {
		return err
	}
	if err = setAccepted(num, id); err != nil {
		return err
	}
	err = memorandums.CheckWifiMemorandumAccepted(id)
	return err
}

func enableOrDisableWifiUser(idString string, num int) error {
	id, err := strconv.Atoi(idString)
	if err != nil {
		return err
	}
	err = setDisabled(num, id)
	return err
}

func saveDepartment(id, departmentId string) (err error) {
	_, err = server.Core.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE id = $2", departmentId, id)
	return
}

func viewWifiUsers(pageString string) (usersList []html.WifiUser, departmentList []html.Department, paging html.Pagination, err error) {
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return nil, nil, html.Pagination{}, err
	}
	count, err := getUserCount()
	if err != nil {
		return nil, nil, html.Pagination{}, err
	}
	departmentList, err = departments.GetAll()
	if err != nil {
		return nil, nil, html.Pagination{}, err
	}
	paging = pagination.Calc(page, count)
	usersList, err = getUserList(paging.PerPage, paging.Offset)
	filterUserNames(usersList)
	return
}

func getUser(id string) (html.WifiUser, []html.Department, error) {
	userId, err := strconv.Atoi(id)
	if err != nil {
		return html.WifiUser{}, nil, err
	}
	user, err := GetWifiUserById(userId)
	if err != nil {
		return html.WifiUser{}, nil, err
	}
	depts, err := departments.GetAll()
	return user, depts, err
}

func updateUser(id, mac, name, phone string) (err error) {
	clearMac, err := check.Mac(mac)
	if err != nil {
		return
	}
	clearName := check.Name(name)
	clearPhone := check.Phone(phone)

	_, err = server.Core.Db.Exec("UPDATE wifiUsers SET mac = $1, username = $2, phonenumber = $3 WHERE id = $4", clearMac, clearName, clearPhone, id)
	return
}