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

func ListMailHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	vars := mux.Vars(r)

	switch vars["action"] {
	case "page":
		paging, memorandums, err := viewMailMemorandums(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.MailMemorandumsPage(memorandums, paging))
	case "accept":
		err := acceptPhoneMemorandum(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
	case "show":
		list, err := getMailMemorandumUsers(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.MailMemorandumPage(list))
	}
}

func viewMailMemorandums(pageString string) (paging html.Pagination, memorandums []html.MailMemorandum, err error) {
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	count, err := getMailCount()
	if err != nil {
		return html.Pagination{}, nil, err
	}
	paging = pagination.Calc(page, count)
	memorandums, err = getMailMemorandums(paging.PerPage, paging.Offset)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}
	return
}

func acceptMailMemorandum(id string) (err error) {
	_, err = server.Core.Db.Exec("UPDATE mailmemorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func getMailMemorandums(limit, offset int) (domains []html.MailMemorandum, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM mailmemorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getMailCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM mailmemorandums")
	return
}

func getMailMemorandumUsers(id string) (list []html.Mail, err error) {
	err = server.Core.Db.Select(&list, "SELECT * FROM mailusers WHERE memorandumid = $1", id)
	return
}
