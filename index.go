package main

import (
	"fmt"
	"log"
	"net/http"

	"encoding/json"
	"errors"
	"net/url"
	"time"

	"git.stingr.net/stingray/kirino_wifi/templates/qtpl_html"
)

func indexHandler(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path == "/favicon.ico" {
		http.ServeFile(w, r, "/static/favicon.ico")
		return
	}
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))

	fmt.Fprint(w, qtpl_html.IndexPage("Доступ к WiFi сети СГУ", isAdmin(r)))
}

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

func checkRecaptcha(ans string) error {
	client := &http.Client{Timeout: 20 * time.Second}
	resp, err := client.PostForm("https://www.google.com/recaptcha/api/siteverify",
		url.Values{"secret": {config.RecaptchaKey}, "response": {ans}})
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
		return errors.New("recaptcha entered incorrect")
	}
	return nil
}
