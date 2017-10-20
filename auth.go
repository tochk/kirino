package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"git.stingr.net/stingray/kirino_wifi/templates/qtpl_html"
	"gopkg.in/ldap.v2"
)

func isAdmin(r *http.Request) bool {
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] != nil {
		return true
	}
	return false
}

func (s *server) adminHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := store.Get(r, "applicationData")
	if session.Values["userName"] != nil {
		http.Redirect(w, r, "/admin/memorandums/", 302)
		return
	}

	fmt.Fprint(w, qtpl_html.AdminPage("Вход в систему"))
}

func auth(login, password string) (string, error) {
	username := ""
	l, err := ldap.Dial("tcp", config.LdapServer)
	if err != nil {
		return username, err
	}
	defer l.Close()

	if l.Bind(config.LdapUser, config.LdapPassword); err != nil {
		return username, err
	}

	searchRequest := ldap.NewSearchRequest(
		config.LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(sAMAccountName="+login+"))",
		[]string{"cn"},
		nil,
	)

	if sr, err := l.Search(searchRequest); err != nil || len(sr.Entries) != 1 {
		err = errors.New("User not found")
		return username, err
	} else {
		username = sr.Entries[0].GetAttributeValue("cn")
	}

	err = l.Bind(username, password)

	return username, err
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	r.ParseForm()
	session, _ := store.Get(r, "applicationData")

	if userName, err := auth(r.Form["login"][0], r.Form["password"][0]); err != nil {
		log.Println(err)
		http.Redirect(w, r, "/admin/", 302)
	} else {
		session, _ = store.Get(r, "applicationData")
		session.Values["userName"] = userName
		session.Save(r, w)
		http.Redirect(w, r, "/admin/memorandums/", 302)
	}
}

func logoutHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := store.Get(r, "applicationData")
	session.Values["userName"] = nil
	session.Save(r, w)
	http.Redirect(w, r, "/admin/", 302)
}
