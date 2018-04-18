package memorandums

import (
	"fmt"
	"log"
	"net/http"

	"github.com/tochk/kirino_wifi/auth"
	"github.com/tochk/kirino_wifi/templates/html"
)

func WifiHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	fmt.Fprint(w, html.IndexPage("Доступ к WiFi сети СГУ", auth.IsAdmin(r)))
}

