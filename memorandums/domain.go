package memorandums

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func ListDomainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var paging html.Pagination
	var memorandums []html.Domain
	var err error
	count, err := getDomainsCount()
	if err != nil {
		log.Println(err)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/domain/memorandums/"):]
	splittedUrl := strings.Split(urlInfo, "/")
	switch splittedUrl[0] {
	case "page":
		page, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		paging = pagination.Calc(page, count)
		memorandums, err = getDomainMemorandums(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println(err)
			return
		}
	case "accept":
		_, err := server.Core.Db.Exec("UPDATE domains SET accepted = 1 WHERE id = $1", splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/admin/domain/memorandums/", 302)
		return
	default:
		if paging.CurrentPage == 0 {
			memorandums, err = getDomainMemorandums(50, 0)
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

	fmt.Fprint(w, html.DomainMemorandumsPage(memorandums, paging))
}

func getDomainMemorandums(limit, offset int) (domains []html.Domain, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM domains ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getDomainsCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM domains")
	return
}
