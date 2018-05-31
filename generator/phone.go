package generator

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/templates/html"
)

func PhoneGenerateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}
	if err := memorandums.CheckRecaptcha(r.FormValue("g-recaptcha-response")); err != nil {
		log.Println(err)
		fmt.Fprint(w, "Капча введена неправильно")
		return
	}
	log.Printf("%#v", r.PostForm)
}

func PhoneGeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/wifi/generated/"):]
	splittedUrl := strings.Split(memorandumInfo, "/")

	var exist []string
	if splittedUrl[1] != "0" {
		exist = strings.Split(splittedUrl[1], ",")
	}

	pageType := "phone"
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	fmt.Fprint(w, html.GeneratedPage(pageType, splittedUrl[0], splittedUrl[2], exist))
}