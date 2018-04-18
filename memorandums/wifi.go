package memorandums

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/tochk/kirino_wifi/templates/html"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		http.ServeFile(w, r, "static/favicon.ico")
		return
	}
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))

	fmt.Fprint(w, html.IndexPage("Доступ к WiFi сети СГУ", isAdmin(r)))
}

