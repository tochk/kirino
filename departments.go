package main

import (
	"html/template"
	"log"
	"net/http"
)

type Department struct {
	Id       int64  `db:"id"`
	Name     string `db:"name"`
	Selected bool
}


func (s *server) departmentsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] == nil {
		http.Redirect(w, r, "/admin/", 302)
		return
	}

	departments, err := s.getDepartments()

	latexTemplate, err := template.ParseFiles("templates/html/departments.tmpl.html")
	if err != nil {
		log.Println(err)
		return
	}
	if err = latexTemplate.Execute(w, departments); err != nil {
		log.Println(err)
		return
	}
}

func (s *server) getDepartments() ([]Department, error) {
	var departments []Department
	err := s.Db.Select(&departments, "SELECT id, left(initcap(name),35) as name FROM departments ORDER BY name ASC")
	return departments, err
}
