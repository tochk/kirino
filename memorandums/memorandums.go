package memorandums

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/mux"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

var (
	RecaptchaError = errors.New("recaptcha entered incorrect")
)

func CheckRecaptcha(ans string) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {server.Config.RecaptchaKey}, "response": {ans}})
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	gr := RecaptchaResponse{Success: false}
	decoder := json.NewDecoder(resp.Body)
	if err = decoder.Decode(&gr); err != nil {
		return err
	}
	if !gr.Success {
		return RecaptchaError
	}
	return nil
}

func FormsHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	vars := mux.Vars(r)
	pageType := vars["type"]
	if auth.IsAdmin(r) {
		pageType = "admin"
	}
	switch vars["type"] {
	case "wifi", "":
		fmt.Fprint(w, html.WifiPage(pageType))
	case "phone":
		fmt.Fprint(w, html.PhonePage(pageType))
	case "mail":
		fmt.Fprint(w, html.MailPage(pageType))
	case "domain":
		fmt.Fprint(w, html.DomainPage(pageType))
	case "ethernet":
		fmt.Fprint(w, html.EthernetPage(pageType))
	}
}
