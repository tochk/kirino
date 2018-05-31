package memorandums

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/templates/html"
)

func PhoneHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if auth.IsAdmin(r) {
		fmt.Fprint(w, html.PhonePage("admin"))
	} else {
		fmt.Fprint(w, html.PhonePage("phone"))
	}
}
