package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
)

var (
	emptySessionKey   = errors.New("empty session key")
	emptyRecaptchaKey = errors.New("empty recaptcha key")
)

var Config struct {
	DbLogin      string `json:"dbLogin"`
	DbPassword   string `json:"dbPassword"`
	DbHost       string `json:"dbHost"`
	DbDb         string `json:"dbDb"`
	DbPort       string `json:"dbPort"`
	LdapUser     string `json:"ldapUser"`
	LdapPassword string `json:"ldapPassword"`
	LdapServer   string `json:"ldapServer"`
	LdapBaseDN   string `json:"ldapBaseDN"`
	SessionKey   string `json:"sessionKey"`
	RecaptchaKey string `json:"recaptchaKey"`
}

var Core struct {
	Db    *sqlx.DB
	Store *sessions.CookieStore
}

func loadConfig(configFile string) error {
	jsonData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, &Config)
	if err != nil {
		return err
	}
	if Config.SessionKey == "" {
		return emptySessionKey
	}
	Core.Store = sessions.NewCookieStore([]byte(Config.SessionKey))
	if Config.RecaptchaKey == "" {
		return emptyRecaptchaKey
	}
	return nil
}
