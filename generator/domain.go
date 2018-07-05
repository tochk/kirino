package generator

import (
	"net/url"

	"github.com/jmoiron/sqlx"
	"github.com/tochk/kirino/check"
	"github.com/tochk/kirino/latex"
	"github.com/tochk/kirino/server"
	"github.com/tochk/kirino/templates/html"
)

func generateDomain(form url.Values) (string, string, error) {
	hash := generateHash(form.Get("nameHost"))

	domain := checkDomainData(html.Domain{
		Hosting:    form.Get("locationHost"),
		FIO:        form.Get("FIOHost"),
		Position:   form.Get("posHost"),
		Accounts:   form.Get("accountHost"),
		Target:     form.Get("target_mail"),
		Department: form.Get("dep1"),
		Name:       form.Get("nameHost"),
	})

	memorandumId, err := writeDomainDataToDb(domain, hash)
	if err != nil {
		return "", "", err
	}

	if err = latex.GenerateDomainMemorandum(domain, hash, memorandumId); err != nil {
		return "", "", err
	}

	return hash, "", nil
}

func checkDomainData(domain html.Domain) html.Domain {
	domain.Hosting = check.All(domain.Hosting)
	domain.FIO = check.All(domain.FIO)
	domain.Position = check.All(domain.Position)
	domain.Accounts = check.All(domain.Accounts)
	domain.Target = check.All(domain.Target)
	domain.Department = check.All(domain.Department)
	domain.Name = check.All(domain.Name)

	return domain
}

func writeDomainDataToDb(data html.Domain, hash string) (int, error) {
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

func tryWriteDomainDataToDb(tx *sqlx.Tx, data html.Domain, hash string) (memorandumId int, err error) {
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
