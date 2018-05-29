package departments

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

type Department = html.Department

func getDepartmentsPagination(limit, offset int) (departments []Department, err error) {
	err = server.Core.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC LIMIT $1 OFFSET $2", limit, offset)
	return
}

func GetAll() (departments []Department, err error) {
	err = server.Core.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return
}

func getCount() (count int, err error) {
	err = server.Core.Db.Select(&count, "SELECT COUNT(*) FROM departments")
	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}

	var (
		departments []Department
		paging      pagination.Pagination
	)

	count, err := getCount()
	if err != nil {
		log.Println(err)
		return
	}

	splittedUrl := strings.Split(r.URL.Path[len("/admin/departments/"):], "/")
	switch splittedUrl[0] {
	case "page":
		page, err := strconv.Atoi(splittedUrl[1])
		if err != nil {
			log.Println(err)
			return
		}

		paging = pagination.Calc(page, count)
		departments, err = getDepartmentsPagination(paging.PerPage, paging.Offset)
		if err != nil {
			log.Println(err)
			return
		}
	case "":
		paging = pagination.Calc(1, count)
		departments, err = getDepartmentsPagination(paging.PerPage, 0)
		if err != nil {
			log.Println(err)
			return
		}

		fmt.Fprint(w, html.DepartmentsPage(departments, paging))
	}
}
