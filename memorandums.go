package main

import (
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
)


type FullWifiMemorandumClientList struct {
	Clients     []FullWifiUser
	Memorandum  FullWifiMemorandum
	Departments []Department
}

type MemorandumsPage struct {
	Memorandums []FullWifiMemorandum
	Departments []Department
	Pagination  Pagination
}

type FullWifiMemorandum struct {
	Id           int    `db:"id"`
	AddTime      string `db:"addtime"`
	Accepted     int    `db:"accepted"`
	Disabled     int    `db:"disabled"`
	DepartmentId *int   `db:"departmentid"`
}

func (s *server) showMemorandumsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var pagination Pagination
	perPage := 50
	var memorandums []FullWifiMemorandum
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/memorandums/"):]
	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "save":
			if len(splittedUrl[1]) > 0 && r.PostForm.Get("department") != "" {
				if _, err := s.Db.Exec("UPDATE memorandums SET departmentid = $1 WHERE id = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
					log.Println(err)
					return
				}
				if _, err := s.Db.Exec("UPDATE wifiUsers SET departmentid = $1 WHERE memorandumid = $2", r.PostForm.Get("department"), splittedUrl[1]); err != nil {
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
			pagination = s.paginationCalc(page, perPage, "memorandums")
			memorandums, err = s.getMemorandums(pagination.PerPage, pagination.Offset)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
	latexTemplate, err := template.ParseFiles("templates/html/memorandums.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}

	if pagination.CurrentPage == 0 {
		memorandums, err = s.getMemorandums(50, 0)
		if err != nil {
			log.Println(err)
			return
		}
		pagination = s.paginationCalc(1, perPage, "memorandums")
	}

	departments, err := s.getDepartments()
	if err != nil {
		log.Println(err)
		return
	}

	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}

	if err = latexTemplate.Execute(w, MemorandumsPage{
		Memorandums: memorandums,
		Departments: departments,
		Pagination:  pagination,
	}); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) acceptMemorandum(id string) (err error) {
	if _, err = s.Db.Exec("UPDATE wifiUsers SET accepted = 1 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = s.Db.Exec("UPDATE memorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func (s *server) rejectMemorandum(id string) (err error) {
	if _, err = s.Db.Exec("UPDATE wifiUsers SET accepted = 2 WHERE memorandumId = $1", id); err != nil {
		return
	}
	_, err = s.Db.Exec("UPDATE memorandums SET accepted = 2 WHERE id = $1", id)
	return
}

func (s *server) checkMemorandumHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := store.Get(r, "applicationData")
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
			if err := s.acceptMemorandum(memId); err != nil {
				log.Println(err)
				return
			}
		} else if memId[0:len("reject/")] == "reject/" {
			memId = memId[len("reject/"):]
			if err := s.rejectMemorandum(memId); err != nil {
				log.Println(err)
				return
			}
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	}
	clientsInMemorandum := make([]FullWifiUser, 0)
	if err := s.Db.Select(&clientsInMemorandum, "SELECT id, mac, userName, phoneNumber, hash, memorandumId, accepted, disabled, departmentid FROM wifiUsers WHERE memorandumId = $1", memId); err != nil {
		log.Println(err)
		return
	}

	for i, e := range clientsInMemorandum {
		if len(e.UserName) > 65 {
			clientsInMemorandum[i].UserName = e.UserName[:66] + "..."
		}
	}

	departments, err := s.getDepartments()
	if err != nil {
		log.Println(err)
		return
	}

	var memorandum FullWifiMemorandum
	if err := s.Db.Get(&memorandum, "SELECT id, addTime, accepted, departmentid FROM memorandums WHERE id = $1", memId); err != nil {
		log.Println(err)
		return
	}

	latexTemplate, err := template.ParseFiles("templates/html/checkMemorandum.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err = latexTemplate.Execute(w, FullWifiMemorandumClientList{Clients: clientsInMemorandum, Memorandum: memorandum, Departments: departments}); err != nil {
		log.Println(err)
		return
	}
}

//todo: rewrite
func (s *server) getMemorandums(limit, offset int) (memorandums []FullWifiMemorandum, err error) {
	err = s.Db.Select(&memorandums, "SELECT id, addTime, accepted, departmentid FROM memorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	var max, min []MemAccepted
	err = s.Db.Select(&max, "SELECT max(accepted) as accepted, memorandumid FROM wifiusers WHERE memorandumid IN (SELECT id FROM memorandums) GROUP BY memorandumid ORDER BY memorandumid desc LIMIT $1 OFFSET $2 ", limit, offset)
	err = s.Db.Select(&min, "SELECT min(accepted) as accepted, memorandumid FROM wifiusers WHERE memorandumid IN (SELECT id FROM memorandums) GROUP BY memorandumid ORDER BY memorandumid desc LIMIT $1 OFFSET $2 ", limit, offset)
	for i, e := range memorandums {
		for _, em := range max {
			for _, emi := range min {
				if e.Id == em.MemorandumId && e.Id == emi.MemorandumId && emi.Accepted != em.Accepted {
					memorandums[i].Accepted = 3
				}
			}
		}
	}
	return
}

type MemAccepted struct {
	MemorandumId int `db:"memorandumid"`
	Accepted     int `db:"accepted"`
}

//todo: rewrite
func (s *server) checkMemorandumAccepted(userId int) error {
	_, err := s.Db.Exec("UPDATE memorandums SET accepted = 1 "+
		"WHERE id = (SELECT memorandumid FROM wifiusers WHERE id = $1) AND "+
		"(SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1))"+
		" - (SELECT COUNT(*) FROM wifiusers WHERE memorandumid = "+
		"(SELECT memorandumid FROM wifiusers WHERE id = $1) AND accepted = 1) = 0;", userId)
	return err
}
