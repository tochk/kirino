package check

import (
	"bytes"
	"errors"
	"regexp"
	"strings"
)

var (
	regForMac       = regexp.MustCompile("[^a-f0-9]+")
	regForName      = regexp.MustCompile("[^а-яА-Яa-zA-Z \\.\\-]+")
	regForPhone     = regexp.MustCompile("[^0-9+\\-() ]+")
	regForAll      = regexp.MustCompile("[^а-яА-Яa-zA-Z0-9 \\.\\-\\_\\:\\;\\,]+")
	InvalidMacError = errors.New("invalid mac-address")
)

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