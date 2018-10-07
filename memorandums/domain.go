package memorandums

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func ListDomainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	vars := mux.Vars(r)

	switch vars["action"] {
	case "page":
		paging, memorandums, err := viewDomainMemorandums(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.DomainMemorandumsPage(memorandums, paging))
	case "accept":
		err := acceptDomainMemorandum(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
	}
}

func viewDomainMemorandums(pageString string) (paging html.Pagination, memorandums []html.Domain, err error) {
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	count, err := getDomainsCount()
	if err != nil {
		return html.Pagination{}, nil, err
	}
	paging = pagination.Calc(page, count)
	memorandums, err = getDomainMemorandums(paging.PerPage, paging.Offset)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}
	return
}

func acceptDomainMemorandum(id string) (err error) {
	_, err = server.Core.Db.Exec("UPDATE domains SET accepted = 1 WHERE id = $1", id)
	return
}

func getDomainMemorandums(limit, offset int) (domains []html.Domain, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM domains ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getDomainsCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM domains")
	return
}
