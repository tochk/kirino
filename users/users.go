package users

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/departments"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type WifiUser = html.WifiUser

func WifiUserHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}

	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/user/"):]
	var user WifiUser
	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "edit":
			if len(splittedUrl[1]) > 0 {
				userId, err := strconv.Atoi(splittedUrl[1])
				if err != nil {
					log.Println(err)
					return
				}
				user, err = GetWifiUserById(userId)
				if err != nil {
					log.Println(err)
					return
				}
			}
		case "save":
			clearMac, err := check.Mac(r.PostForm.Get("mac1"))
			if err != nil {
				log.Println(err)
				return
			}
			clearName := check.Name(r.PostForm.Get("user1"))
			clearPhone := check.Phone(r.PostForm.Get("tel1"))

			_, err = server.Core.Db.Exec("UPDATE wifiUsers SET mac = $1, username = $2, phonenumber = $3 WHERE id = $4", clearMac, clearName, clearPhone, splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			http.Redirect(w, r, "/admin/wifi/users/", 302)
			return
		}
	}

	depts, err := departments.GetAll()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, html.WifiUserPage(user, depts))
}

func getUserList(limit, offset int) (userList []WifiUser, err error) {
	err = server.Core.Db.Select(&userList, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getUserCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM wifiUsers")
	return
}

func GetWifiUserById(id int) (user WifiUser, err error) {
	err = server.Core.Db.Get(&user, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers WHERE id = $1", id)
	return
}

func GetWifiUserByMac(mac string) (user WifiUser, err error) {
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

func WifiUsersHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/wifi/users/"):]
	var (
		usersList []WifiUser
		paging    pagination.Pagination
		err       error
	)
	count, err := getUserCount()
	if err != nil {
		log.Println(err)
		return
	}

	splittedUrl := strings.Split(urlInfo, "/")
	switch splittedUrl[0] {
	case "savedept":
		if len(splittedUrl[1]) > 0 && r.PostForm.Get("department") != "" {
			if _, err := server.Core.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE id = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
				log.Println(err)
				return
			}
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "accept":
		id, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		if err = setAccepted(1, id); err != nil {
			log.Println(err)
			return
		}
		if err = memorandums.CheckMemorandumAccepted(id); err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "reject":
		id, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		if err = setAccepted(2, id); err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "enable":
		id, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		if err = setDisabled(0, id); err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "disable":
		id, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		if err = setDisabled(1, id); err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "page":
		page, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}

		paging = pagination.Calc(page, count)
		usersList, err = getUserList(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println(err)
			return
		}
	case "search":
		var err error
		usersList, err = getSearchResult(r.URL.Query())
		if err != nil {
			log.Println(err)
			return
		}
	default:
		usersList, err = getUserList(50, 0)
		if err != nil {
			log.Println(err)
			return
		}
		paging = pagination.Calc(1, count)
	}

	for i, e := range usersList {
		if len(e.UserName) > 65 {
			usersList[i].UserName = e.UserName[:66] + "..."
		}
	}

	depts, err := departments.GetAll()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, html.WifiUsersPage(usersList, depts, paging))
}

//todo search
func getSearchResult(values url.Values) (userList []WifiUser, err error) {
	err = server.Core.Db.Select(&userList, "SELECT id, mac, userName, phoneNumber, accepted, disabled, departmentid, memorandumId FROM wifiUsers WHERE mac LIKE CONCAT(CONCAT('%', $1), '%') AND username LIKE CONCAT(CONCAT('%', $2), '%') AND phonenumber LIKE CONCAT(CONCAT('%', $3), '%') ORDER BY id DESC ", values.Get("mac"), values.Get("name"), values.Get("phone"))
	return
}
