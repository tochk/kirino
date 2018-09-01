package memorandums

import (
	"fmt"
	log "github.com/Sirupsen/logrus"
	"net/http"
	"strconv"
	"strings"

	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func ListMailHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var paging html.Pagination
	var memorandums []html.MailMemorandum
	var err error
	count, err := getMailCount()
	if err != nil {
		log.Println(err)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/mail/memorandums/"):]
	splittedUrl := strings.Split(urlInfo, "/")
	switch splittedUrl[0] {
	case "page":
		page, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		paging = pagination.Calc(page, count)
		memorandums, err = getMailMemorandums(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println(err)
			return
		}
	case "accept":
		_, err := server.Core.Db.Exec("UPDATE mailmemorandums SET accepted = 1 WHERE id = $1", splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/admin/mail/memorandums/", 302)
		return
	default:
		if paging.CurrentPage == 0 {
			memorandums, err = getMailMemorandums(50, 0)
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

	fmt.Fprint(w, html.MailMemorandumsPage(memorandums, paging))
}

func getMailMemorandums(limit, offset int) (domains []html.MailMemorandum, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM mailmemorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getMailCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM mailmemorandums")
	return
}

func ViewMailHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	memId := r.URL.Path[len("/admin/mail/memorandum/"):]
	if memId == "" {
		log.Println("Invalid memorandum id")
		return
	}

	list, err := getMailMemorandumUsers(memId)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, html.MailMemorandumPage(list))
}

func getMailMemorandumUsers(id string) (list []html.Mail, err error) {
	err = server.Core.Db.Select(&list, "SELECT * FROM mailusers WHERE memorandumid = $1", id)
	return
}
