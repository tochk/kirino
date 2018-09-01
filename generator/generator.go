package generator

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"

	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/templates/html"
)

func generateHash(word string) string {
	hasher := sha256.New()
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)
	hashStr := word + strconv.Itoa(r1.Intn(1000000))
	hasher.Write([]byte(hashStr))
	hashedStr := hex.EncodeToString(hasher.Sum(nil))
	if file, err := os.Open("userFiles\\" + hashStr + ".tex"); err == nil {
		file.Close()
		hashedStr = generateHash(hashedStr)
	}
	return hashedStr
}

func GenerateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if err := r.ParseForm(); err != nil {
		log.Println(err)
		return
	}
	if err := check.Recaptcha(r.FormValue("g-recaptcha-response")); err != nil {
		log.Println(err)
		fmt.Fprint(w, "Капча введена неправильно")
		return
	}
	vars := mux.Vars(r)

	var (
		hash, getData string
		err           error
	)
	switch vars["type"] {
	case "wifi":
		hash, getData, err = generateWifi(r.PostForm)
	case "domain":
		hash, getData, err = generateDomain(r.PostForm)
	case "ethernet":
		hash, getData, err = generateEthernet(r.PostForm)
	case "mail":
		hash, getData, err = generateMail(r.PostForm)
	case "phone":
		hash, getData, err = generatePhone(r.PostForm)
	}
	if err != nil {
		fmt.Fprint(w, html.ErrorPage(auth.IsAdmin(r), err))
		log.Print("Error on generate page: ", err)
		return
	}
	http.Redirect(w, r, "/generated/"+vars["type"]+"/"+hash+"/"+getData, 302)
}

func GeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	vars := mux.Vars(r)
	pageType := vars["type"]
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	switch vars["type"] {
	case "wifi":
		var exist []string
		if r.URL.Query().Get("exist") != "0" {
			exist = strings.Split(r.URL.Query().Get("exist"), ",")
		}
		fmt.Fprint(w, html.GeneratedPage(pageType, vars["token"], r.URL.Query().Get("count"), exist))
	case "ethernet":
		fmt.Fprint(w, html.EthernetGeneratedPage(pageType, vars["token"]))
	case "phone":
		fmt.Fprint(w, html.PhoneGeneratedPage(pageType, vars["token"]))
	case "domain":
		fmt.Fprint(w, html.DomainGeneratedPage(pageType, vars["token"]))
	case "mail":
		fmt.Fprint(w, html.MailGeneratedPage(pageType, vars["token"]))
	}
}
