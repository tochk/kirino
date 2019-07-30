package departments

import (
	"fmt"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/pagination"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}

	count, err := getCount()
	if err != nil {
		log.Println("Error on departments page: ", err)
		fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
		return
	}

	var paging html.Pagination

	vars := mux.Vars(r)
	switch vars["action"] {
	case "view":
		page, err := strconv.Atoi(vars["num"])
		if err != nil {
			log.Println("Error on departments page: ", err)
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			return
		}
		paging = pagination.Calc(page, count)

		departments, err := getDepartmentsPagination(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println("Error on departments page: ", err)
			fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
			return
		}
		fmt.Fprint(w, html.DepartmentsPage(departments, paging))
	}
}

func GetAll() (departments []html.Department, err error) {
	err = server.Core.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return
}

func getDepartmentsPagination(limit, offset int) (departments []html.Department, err error) {
	err = server.Core.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC LIMIT $1 OFFSET $2", limit, offset)
	return
}

func getCount() (count int, err error) {
	err = server.Core.Db.Get(&count, "SELECT COUNT(*) FROM departments")
	return
}
