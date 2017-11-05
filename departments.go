package main

import (
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"git.stingr.net/stingray/kirino_wifi/templates/qtpl_html"
)

type Department = qtpl_html.Department

func (s *server) departmentsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !isAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}
	var (
		urlInfo     = r.URL.Path[len("/admin/departments/"):]
		perPage     = 50
		pagination  Pagination
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
			pagination = s.paginationCalc(page, perPage, "departments")
			departments, err = s.getDepartments(pagination.PerPage, pagination.Offset)
			if err != nil {
				log.Println(err)
				return
			}
		}
	}
	if len(departments) == 0 {
		departments, err = s.getDepartments(perPage, 0)
		if err != nil {
			log.Println(err)
			return
		}
	}

	fmt.Fprint(w, qtpl_html.DepartmentsPage("Подразделения", departments, pagination))
}

func (s *server) getDepartments(limit, offset int) (departments []Department, err error) {
	err = s.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC LIMIT $1 OFFSET $2", limit, offset)
	return
}

func (s *server) getAllDepartments() (departments []Department, err error) {
	err = s.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return
}
