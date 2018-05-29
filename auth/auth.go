package auth

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
	"gopkg.in/ldap.v2"
)

var (
	userNotFoundError  = errors.New("user not found")
	emptyPasswordError = errors.New("empty password")
)

func IsAdmin(r *http.Request) bool {
	session, _ := server.Core.Store.Get(r, "kirino_session")
	if session.Values["userName"] != nil {
		return true
	}
	return false
}

func auth(login, password string) (string, error) {
	if password == "" {
		return "", emptyPasswordError
	}
	username := ""
	l, err := ldap.Dial("tcp", server.Config.LdapServer)
	if err != nil {
		return "", err
	}
	l.Close()

	l, err = ldap.Dial("tcp", server.Config.LdapServer)
	if err != nil {
		return "", err
	}
	defer l.Close()
	if l.Bind(server.Config.LdapUser, server.Config.LdapPassword); err != nil {
		return "", err
	}

	searchRequest := ldap.NewSearchRequest(
		server.Config.LdapBaseDN,
		ldap.ScopeWholeSubtree, ldap.NeverDerefAliases, 0, 0, false,
		"(&(sAMAccountName="+login+"))",
		[]string{"cn"},
		nil,
	)

	if sr, err := l.Search(searchRequest); err != nil || len(sr.Entries) != 1 {
		return username, userNotFoundError
	} else {
		username = sr.Entries[0].GetAttributeValue("cn")
	}

	if err = l.Bind(username, password); err != nil {
		return "", err
	}
	return username, err
}

func Handler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	session, _ := server.Core.Store.Get(r, "kirino_session")
	url := strings.Split(r.URL.Path, "/")
	switch url[2] {
	case "login":
		if r.Method == "GET" {
			if IsAdmin(r) {
				http.Redirect(w, r, "/admin/memorandums/", 302)
				return
			}
			fmt.Fprint(w, html.AdminPage())
		} else {
			r.ParseForm()
			session, _ := server.Core.Store.Get(r, "kirino_session")
			if userName, err := auth(r.Form["login"][0], r.Form["password"][0]); err != nil {
				log.Println(err)
				http.Redirect(w, r, "/admin/", 302)
			} else {
				session, _ = server.Core.Store.Get(r, "kirino_session")
				session.Values["userName"] = userName
				session.Save(r, w)
				http.Redirect(w, r, "/admin/memorandums/", 302)
			}
		}
	case "logout":
		session.Values["userName"] = nil
		session.Save(r, w)
		http.Redirect(w, r, "/admin/", 302)
	}
}
