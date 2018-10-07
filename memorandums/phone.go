package memorandums

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"

	"net/http"
	"strconv"
	"strings"

	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func ListPhoneHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	vars := mux.Vars(r)

	switch vars["action"] {
	case "view":
		paging, memorandums, err := viewPhoneMemorandums(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.PhoneMemorandumsPage(memorandums, paging))
	case "accept":
		err := acceptPhoneMemorandum(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
	case "show":
		list, err := getPhoneMemorandumUsers(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.PhoneMemorandumPage(list))
	}
}

func viewPhoneMemorandums(pageString string) (paging html.Pagination, memorandums []html.PhoneMemorandum, err error) {
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	count, err := getPhoneCount()
	if err != nil {
		return html.Pagination{}, nil, err
	}
	paging = pagination.Calc(page, count)
	memorandums, err = getPhoneMemorandums(paging.PerPage, paging.Offset)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}
	return
}

func acceptPhoneMemorandum(id string) (err error) {
	_, err = server.Core.Db.Exec("UPDATE phonememorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func getPhoneMemorandums(limit, offset int) (domains []html.PhoneMemorandum, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM phonememorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getPhoneCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM phonememorandums")
	return
}

func getPhoneMemorandumUsers(id string) (list []html.Phone, err error) {
	err = server.Core.Db.Select(&list, "SELECT * FROM phoneusers WHERE memorandumid = $1", id)
	return
}
