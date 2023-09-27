package sql

import (
	"context"
	"database/sql"
	"ronce/src/go/log"

	"github.com/jmoiron/sqlx"
)

type Tx struct {
	tx     *sqlx.Tx
	logger *log.Logger
	debug  bool
}

func (db *DB) Begin(ctx context.Context) (*Tx, error) {
	tx, err := db.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &Tx{tx: tx, logger: db.Logger, debug: db.Debug}, nil
}

func (tx *Tx) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return exec(ctx, tx.tx, tx.logger, tx.debug, query, args...)
}

func (tx *Tx) Get(ctx context.Context, into any, query string, args ...any) error {
	return get(ctx, tx.tx, tx.logger, tx.debug, into, query, args...)
}

func (tx *Tx) Select(ctx context.Context, into any, query string, args ...any) error {
	return selectx(ctx, tx.tx, tx.logger, tx.debug, into, query, args...)
}

func (tx *Tx) Commit() error {
	return tx.tx.Commit()
}

func (tx *Tx) Rollback() error {
	return tx.tx.Rollback()
}
