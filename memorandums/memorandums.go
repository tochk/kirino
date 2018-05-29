package memorandums

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/tochk/kirino/departments"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type WifiMemorandum = html.Memorandum
type WifiUser = html.WifiUser

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

type MemAccepted struct {
	MemorandumId int `db:"memorandumid"`
	MaxAccepted  int `db:"maxAccepted"`
	MinAccepted  int `db:"minAccepted"`
}

var (
	RecaptchaError = errors.New("recaptcha entered incorrect")
)

func ListWifiHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var paging pagination.Pagination
	var memorandums []WifiMemorandum
	var err error
	count, err := getWifiUserCount()
	if err != nil {
		log.Println(err)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/memorandums/"):]
	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "save":
			if len(splittedUrl[1]) > 0 && r.PostForm.Get("department") != "" {
				if _, err := server.Core.Db.Exec("UPDATE memorandums SET departmentid = $1 WHERE id = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
					log.Println(err)
					return
				}
				if _, err := server.Core.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE memorandumid = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
					log.Println(err)
					return
				}
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
			memorandums, err = getMemorandums(paging.PerPage, paging.Offset)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}

	if paging.CurrentPage == 0 {
		memorandums, err = getMemorandums(50, 0)
		if err != nil {
			log.Println(err)
			return
		}
		paging = pagination.Calc(1, count)
	}

	departmentList, err := departments.GetAll()
	if err != nil {
		log.Println(err)
		return
	}

	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}

	fmt.Fprint(w, html.WifiMemorandumsPage(memorandums, departmentList, paging))
}

func acceptMemorandum(id string) (err error) {
	if _, err = server.Core.Db.Exec("UPDATE wifiUsers SET accepted = 1 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = server.Core.Db.Exec("UPDATE memorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func rejectMemorandum(id string) (err error) {
	if _, err = server.Core.Db.Exec("UPDATE wifiUsers SET accepted = 2 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = server.Core.Db.Exec("UPDATE memorandums SET accepted = 2 WHERE id = $1", id)
	return
}

func getWifiUserCount() (count int, err error) {
	err = server.Core.Db.Select(&count, "SELECT COUNT(*) FROM wifiUsers")
	return
}

func ViewWifiHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	memId := r.URL.Path[len("/admin/checkMemorandum/"):]
	if memId == "" {
		log.Println("Invalid memorandum id")
		return
	}

	if len(memId) > len("accept/") {
		if memId[0:len("accept/")] == "accept/" {
			memId = memId[len("accept/"):]
			if err := acceptMemorandum(memId); err != nil {
				log.Println(err)
				return
			}
		} else if memId[0:len("reject/")] == "reject/" {
			memId = memId[len("reject/"):]
			if err := rejectMemorandum(memId); err != nil {
				log.Println(err)
				return
			}
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	}
	clientsInMemorandum := make([]WifiUser, 0)
	if err := server.Core.Db.Select(&clientsInMemorandum, "SELECT id, mac, userName, phoneNumber, hash, memorandumId, accepted, disabled, departmentid FROM wifiUsers WHERE memorandumId = $1", memId); err != nil {
		log.Println(err)
		return
	}

	for i, e := range clientsInMemorandum {
		if len(e.UserName) > 65 {
			clientsInMemorandum[i].UserName = e.UserName[:66] + "..."
		}
	}

	departmentList, err := departments.GetAll()
	if err != nil {
		log.Println(err)
		return
	}

	var memorandum WifiMemorandum
	if err := server.Core.Db.Get(&memorandum, "SELECT id, addTime, accepted, departmentid FROM memorandums WHERE id = $1", memId); err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, html.WifiMemorandumPage(memorandum, clientsInMemorandum, departmentList))
}

func getMemorandums(limit, offset int) (memorandums []WifiMemorandum, err error) {
	if err = server.Core.Db.Select(&memorandums, "SELECT id, addTime, accepted, departmentid FROM memorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset); err != nil {
		return
	}
	var accepted []MemAccepted
	err = server.Core.Db.Select(&accepted, "SELECT accepted(accepted) as maxAccepted, min(accepted) as minAccepted memorandumid FROM wifiusers WHERE memorandumid IN (SELECT id FROM memorandums) GROUP BY memorandumid ORDER BY memorandumid desc LIMIT $1 OFFSET $2 ", limit, offset)
	for i, e := range memorandums {
		for _, em := range accepted {
			if e.Id == em.MemorandumId && em.MaxAccepted != em.MinAccepted {
				memorandums[i].Accepted = 3
			}
		}
	}
	return
}

//todo rewrite
func CheckMemorandumAccepted(userId int) error {
	_, err := server.Core.Db.Exec("UPDATE memorandums SET accepted = 1 "+
		"WHERE id = (SELECT memorandumid FROM wifiusers WHERE id = $1) AND "+
		"(SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1))"+
		" - (SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1) AND accepted = 1) = 0;", userId)
	return err
}

func CheckRecaptcha(ans string) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {server.Config.RecaptchaKey}, "response": {ans}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	gr := RecaptchaResponse{Success: false}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&gr); err != nil {
		return err
	}
	if !gr.Success {
		return RecaptchaError
	}
	return nil
}
