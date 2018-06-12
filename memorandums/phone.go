package memorandums

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type PhoneMemorandum = html.PhoneMemorandum
type Phone = html.Phone

func ListPhoneHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var paging pagination.Pagination
	var memorandums []PhoneMemorandum
	var err error
	count, err := getPhoneCount()
	if err != nil {
		log.Println(err)
		return
	}
	r.ParseForm()
	urlInfo := r.URL.Path[len("/admin/phone/memorandums/"):]
	splittedUrl := strings.Split(urlInfo, "/")
	switch splittedUrl[0] {
	case "page":
		page, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		paging = pagination.Calc(page, count)
		memorandums, err = getPhoneMemorandums(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println(err)
			return
		}
	case "accept":
		_, err := server.Core.Db.Exec("UPDATE phonememorandums SET accepted = 1 WHERE id = $1", splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}
		http.Redirect(w, r, "/admin/phone/memorandums/", 302)
		return
	default:
		if paging.CurrentPage == 0 {
			memorandums, err = getPhoneMemorandums(50, 0)
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

	fmt.Fprint(w, html.PhoneMemorandumsPage(memorandums, paging))
}

func getPhoneMemorandums(limit, offset int) (domains []PhoneMemorandum, err error) {
	err = server.Core.Db.Select(&domains, "SELECT * FROM phonememorandums ORDER BY id DESC LIMIT $1 OFFSET $2 ", limit, offset)
	return
}

func getPhoneCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM phonememorandums")
	return
}

func ViewPhoneHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	memId := r.URL.Path[len("/admin/phone/memorandum/"):]
	if memId == "" {
		log.Println("Invalid memorandum id")
		return
	}

	list, err := getPhoneMemorandumUsers(memId)
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, html.PhoneMemorandumPage(list))
}

func getPhoneMemorandumUsers(id string) (list []Phone, err error) {
	err = server.Core.Db.Select(&list, "SELECT * FROM phoneusers WHERE memorandumid = $1", id)
	return
}
