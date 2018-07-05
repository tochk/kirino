package check

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/tochk/kirino/server"
)

var (
	regForMac       = regexp.MustCompile("[^a-f0-9]+")
	regForName      = regexp.MustCompile("[^а-яА-Яa-zA-Z \\.\\-]+")
	regForPhone     = regexp.MustCompile("[^0-9+\\-() ]+")
	regForAll       = regexp.MustCompile("[^а-яА-Яa-zA-Z0-9 \\.\\-\\_\\:\\;\\,@]+")
	InvalidMacError = errors.New("invalid mac-address")
	RecaptchaError  = errors.New("recaptcha entered incorrect")
)

type RecaptchaResponse struct {
	Success bool `json:"success"`
}

func Mac(mac string) (string, error) {
	mac = string(bytes.ToLower([]byte(mac)))
	mac = regForMac.ReplaceAllString(mac, "")
	if len(mac) != 12 {
		return "", InvalidMacError
	}
	return mac, nil
}

func Name(name string) string {
	name = strings.Replace(name, "Ё", "Е", -1)
	name = strings.Replace(name, "ё", "е", -1)
	return regForName.ReplaceAllString(name, "")
}

func Phone(phone string) string {
	return regForPhone.ReplaceAllString(phone, "")
}

func All(str string) string {
	return regForAll.ReplaceAllString(str, "")
}

func Recaptcha(ans string) error {
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
