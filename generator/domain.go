package generator

import (
	"fmt"
	"log"
	"net/http"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/auth"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/memorandums"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func DomainGenerateHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	if err := r.ParseForm(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := memorandums.CheckRecaptcha(r.FormValue("g-recaptcha-response")); err != nil {
		log.Println(err)
		fmt.Fprint(w, "Капча введена неправильно")
		return
	}

	hash := generateHash(r.PostFormValue("nameHost"))

	domain := checkDomainData(Domain{
		Hosting:    r.PostForm.Get("locationHost"),
		FIO:        r.PostForm.Get("FIOHost"),
		Position:   r.PostForm.Get("posHost"),
		Accounts:   r.PostForm.Get("accountHost"),
		Target:     r.PostForm.Get("target_mail"),
		Department: r.PostForm.Get("dep1"),
		Name:       r.PostForm.Get("nameHost"),
	})

	memorandumId, err := writeDomainDataToDb(domain, hash)
	if err != nil {
		log.Println(err)
		return
	}

	if err = latex.GenerateDomainMemorandum(domain, hash, memorandumId); err != nil {
		log.Println(err)
		return
	}

	http.Redirect(w, r, "/domain/generated/"+hash, 302)
}

func checkDomainData(domain Domain) Domain {
	domain.Hosting = check.All(domain.Hosting)
	domain.FIO = check.All(domain.FIO)
	domain.Position = check.All(domain.Position)
	domain.Accounts = check.All(domain.Accounts)
	domain.Target = check.All(domain.Target)
	domain.Department = check.All(domain.Department)
	domain.Name = check.All(domain.Name)

	return domain
}

func writeDomainDataToDb(data Domain, hash string) (int, error) {
	tx, err := server.Core.Db.Beginx()
	if err != nil {
		return 0, err
	}
	id, err := tryWriteDomainDataToDb(tx, data, hash)
	if err != nil {
		tx.Rollback()
		return 0, err
	}
	if err = tx.Commit(); err == nil {
		return id, nil
	}
	return 0, nil
}

func tryWriteDomainDataToDb(tx *sqlx.Tx, data Domain, hash string) (memorandumId int, err error) {
	if err = tx.Get(&memorandumId, "SELECT max(id) FROM domains"); err != nil {
		if err.Error() == "sql: Scan error on column index 0: converting driver.Value type <nil> (\"<nil>\") to a int: invalid syntax" {
			memorandumId = 0
		} else {
			return 0, err
		}
	}
	memorandumId++
	if _, err := tx.Exec(tx.Rebind("INSERT INTO domains (id, addtime, department, name, host, username, hash) VALUES (?, current_date(), ?, ?, ?, ?, ?)"),
		memorandumId, data.Department, data.Name+"|||"+data.Position+"|||"+data.Accounts, data.Hosting, data.FIO, hash); err != nil {
		return 0, err
	}

	return memorandumId, nil
}

func DomainGeneratedHandler(w http.ResponseWriter, r *http.Request) {
	log.Printf("Loaded %s page from %s", r.URL.Path, r.Header.Get("X-Real-IP"))
	memorandumInfo := r.URL.Path[len("/domain/generated/"):]

	pageType := "domain"
	if auth.IsAdmin(r) {
		pageType = "admin"
	}

	fmt.Fprint(w, html.DomainGeneratedPage(pageType, memorandumInfo))
}
