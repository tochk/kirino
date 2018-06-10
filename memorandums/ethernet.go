package memorandums

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type EthernetMemorandum = html.EthernetMemorandum
type Ethernet = html.Ethernet

func EthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.EthernetPage("admin"))
	} else {
		fmt.Fprint(w, html.EthernetPage("ethernet"))
	}
}

func ListEthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var paging pagination.Pagination
	var memorandums []EthernetMemorandum
	var err error
	count, err := getEthernetCount()
	if err != nil {
		log.Println(err)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/ethernet/memorandums/"):]
	splittedUrl := strings.Split(urlInfo, "/")
	switch splittedUrl[0] {
	case "page":
		page, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		paging = pagination.Calc(page, count)
		memorandums, err = getEthernetMemorandums(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println(err)
			return
		}
	case "accept":
		_, err := server.Core.Db.Exec("UPDATE ethmemorandums SET accepted = 1 WHERE id = $1", splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/admin/ethernet/memorandums/", 302)
		return
	default:
		if paging.CurrentPage == 0 {
			memorandums, err = getEthernetMemorandums(50, 0)
			if err != nil {
				log.Println(err)
				return
			}
			paging = pagination.Calc(1, count)
		}
	}

	for index, memorandum := range memorandums {
		memorandums[index].AddTime = strings.Split(memorandum.AddTime, "T")[0]
	}

	fmt.Fprint(w, html.EthernetMemorandumsPage(memorandums, paging))
}

func getEthernetMemorandums(limit, offset int) (domains []EthernetMemorandum, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM ethmemorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getEthernetCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM ethmemorandums")
	return
}


func ViewEthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	memId := r.URL.Path[len("/admin/ethernet/memorandum/"):]
	if memId == "" {
		log.Println("Invalid memorandum id")
		return
	}

	list, err := getEthernetMemorandumUsers(memId)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, html.EthernetMemorandumPage(list))
}

func getEthernetMemorandumUsers(id string) (list []Ethernet, err error) {
	err = server.Core.Db.Select(&list, "SELECT * FROM ethusers WHERE memorandumid = $1", id)
	return
}