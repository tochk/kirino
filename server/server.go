package server

import (
	"encoding/json"
	"errors"
	"io/ioutil"

	"github.com/gorilla/sessions"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

var (
	EmptySessionKey   = errors.New("empty session key")
	EmptyRecaptchaKey = errors.New("empty recaptcha key")
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
	PerPage      int    `json:"perPage"`
}

var Core struct {
	Db    *sqlx.DB
	Store *sessions.CookieStore
}

func LoadConfig(configFile string) error {
	jsonData, err := ioutil.ReadFile(configFile)
	if err != nil {
		return err
	}
	err = json.Unmarshal(jsonData, &Config)
	if err != nil {
		return err
	}
	if Config.SessionKey == "" {
		return EmptySessionKey
	}
	Core.Store = sessions.NewCookieStore([]byte(Config.SessionKey))
	if Config.RecaptchaKey == "" {
		return EmptyRecaptchaKey
	}
	if Config.PerPage == 0 {
		Config.PerPage = 50
	}
	return nil
}

func ConnectToDb() {
	Core.Db = sqlx.MustConnect("postgres", "host="+Config.DbHost+" port="+Config.DbPort+" user="+Config.DbLogin+" dbname="+Config.DbDb+" password="+Config.DbPassword)
}
