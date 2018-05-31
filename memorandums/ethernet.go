package memorandums

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/templates/html"
)

func EthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.EthernetPage("admin"))
	} else {
		fmt.Fprint(w, html.EthernetPage("ethernet"))
	}
}

func ListEthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.EthernetPage("admin"))
	} else {
		fmt.Fprint(w, html.EthernetPage("ethernet"))
	}
}

func ViewEthernetHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.EthernetPage("admin"))
	} else {
		fmt.Fprint(w, html.EthernetPage("ethernet"))
	}
}
