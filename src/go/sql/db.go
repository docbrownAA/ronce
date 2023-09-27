package sql

import (
	"context"
	"database/sql"
	"ronce/src/go/app"
	"ronce/src/go/log"
	"strings"

	"github.com/jmoiron/sqlx"
)

// DB is a wrapper around sqlx.DB for compatibility with synthesio/zconfig.
type DB struct {
	db           *sqlx.DB
	Logger       *log.Logger `inject:"logger"`
	DSN          string      `key:"dsn"            description:"data string connection (host=<> user=<> password=<> dbname=<>)"`
	Debug        bool        `key:"debug"          description:"activate debug logs"`
	MaxOpenConns uint        `key:"max-open-conns" description:"maximum number of connections in the pool"`
	MaxIdleConns uint        `key:"max-idle-conns" description:"maximum number of idle connections in the pool"`
}

func (db *DB) Init() error {
	// Automatically add an appplication name in the DSN so the
	// administration interfaces of Postgres can tell who is connected.
	if !strings.Contains(db.DSN, "application_name") {
		db.DSN += " application_name=" + app.Name + "-" + app.Version
	}

	var err error
	db.db, err = sqlx.Connect("postgres", db.DSN)
	if err != nil {
		return err
	}
	db.db.SetMaxOpenConns(int(db.MaxOpenConns))
	db.db.SetMaxIdleConns(int(db.MaxIdleConns))
	db.Logger = db.Logger.With("module", "sql")
	return nil
}

func (db *DB) Close() error {
	return db.db.Close()
}

func (db *DB) Select(ctx context.Context, into any, query string, args ...any) error {
	return selectx(ctx, db.db, db.Logger, db.Debug, into, query, args...)
}

func (db *DB) Get(ctx context.Context, into any, query string, args ...any) error {
	return get(ctx, db.db, db.Logger, db.Debug, into, query, args...)
}

func (db *DB) Exec(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return exec(ctx, db.db, db.Logger, db.Debug, query, args...)
}
