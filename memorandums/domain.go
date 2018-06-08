package memorandums

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/templates/html"
)

func DomainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.DomainPage("admin"))
	} else {
		fmt.Fprint(w, html.DomainPage("domain"))
	}
}

func ListDomainHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.DomainPage("admin"))
	} else {
		fmt.Fprint(w, html.DomainPage("domain"))
	}
}