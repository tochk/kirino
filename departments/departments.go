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

func GetDepartmentsPagination(limit, offset int) (departments []Department, err error) {
	err = server.Core.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC LIMIT $1 OFFSET $2", limit, offset)
	return
}

func GetDepartments() (departments []Department, err error) {
	err = server.Core.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !auth.IsAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var (
		urlInfo     = r.URL.Path[len("/admin/departments/"):]
		perPage     = 50
		paging      pagination.Pagination
		departments []Department
		err         error
	)

	if len(urlInfo) > 0 {
		splittedUrl := strings.Split(urlInfo, "/")
		switch splittedUrl[0] {
		case "page":
			page, err := strconv.Atoi(splittedUrl[1])
			if err != nil {
				log.Println(err)
				return
			}
			paging = pagination.Calc(page, perPage, "departments")
			departments, err = GetDepartmentsPagination(pagination.PerPage, pagination.Offset)
			if err != nil {
				log.Println(err)
				return
			}
		default:
			paging = s.paginationCalc(1, perPage, "departments")
		}
	} else {
		paging = s.paginationCalc(1, perPage, "departments")
	}

	if len(departments) == 0 {
		departments, err = s.getDepartments(perPage, 0)
		if err != nil {
			log.Println(err)
			return
		}
	}
	fmt.Fprint(w, html.DepartmentsPage("Подразделения", departments, paging))
}
