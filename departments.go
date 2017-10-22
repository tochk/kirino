package main

import (
	"fmt"
	"log"
	"net/http"

	"git.stingr.net/stingray/kirino_wifi/templates/qtpl_html"
)

type Department = qtpl_html.Department

func (s *server) departmentsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if !isAdmin(r) {
		http.Redirect(w, r, "/admin/", 302)
		return
	}

	departments, err := s.getDepartments()
	if err != nil {
		log.Println(err)
		return
	}

	fmt.Fprint(w, qtpl_html.DepartmentsPage("Подразделения", departments))

}

func (s *server) getDepartments() (departments []Department, err error) {
	err = s.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return
}
