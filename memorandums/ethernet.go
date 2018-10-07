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

func ListEthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	vars := mux.Vars(r)

	switch vars["action"] {
	case "view":
		paging, memorandums, err := viewEthernetMemorandums(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.EthernetMemorandumsPage(memorandums, paging))
	case "accept":
		err := acceptEthernetMemorandum(vars["id"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		http.Redirect(w, r, r.Referer(), 302)
		return
	case "show":
		list, err := getEthernetMemorandumUsers(vars["num"])
		if err != nil {
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			log.Print(err)
			return
		}
		fmt.Fprint(w, html.EthernetMemorandumPage(list))
	}
}

func viewEthernetMemorandums(pageString string) (paging html.Pagination, memorandums []html.EthernetMemorandum, err error) {
	page, err := strconv.Atoi(pageString)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	count, err := getEthernetCount()
	if err != nil {
		return html.Pagination{}, nil, err
	}
	paging = pagination.Calc(page, count)
	memorandums, err = getEthernetMemorandums(paging.PerPage, paging.Offset)
	if err != nil {
		return html.Pagination{}, nil, err
	}
	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}
	return
}

func acceptEthernetMemorandum(id string) (err error) {
	_, err = server.Core.Db.Exec("UPDATE ethmemorandums SET accepted = 1 WHERE id = $1", id)
	return
}

func getEthernetMemorandums(limit, offset int) (domains []html.EthernetMemorandum, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM ethmemorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getEthernetCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM ethmemorandums")
	return
}

func getEthernetMemorandumUsers(id string) (list []html.Ethernet, err error) {
	err = server.Core.Db.Select(&list, "SELECT * FROM ethusers WHERE memorandumid = $1", id)
	return
}
