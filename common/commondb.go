package common

import (
	"context"

	"github.com/jmoiron/sqlx"
)

func RunTx(ctx context.Context, db *sqlx.DB, f func(*sqlx.Tx) error) error {
	tx, err := db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
		}
		err = tx.Commit()
	}()
	err = f(tx)
	return err
}
