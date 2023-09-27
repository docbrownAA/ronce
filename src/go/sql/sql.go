package sql

import (
	"context"
	"database/sql"
	"time"

	"ronce/src/go/log"
)

// Queryer is the query runner interface. It is implemented by DB and TX.
type Queryer interface {
	Exec(ctx context.Context, query string, args ...any) (sql.Result, error)
	Get(ctx context.Context, into any, query string, args ...any) error
	Select(ctx context.Context, into any, query string, args ...any) error
}

type queryer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	GetContext(ctx context.Context, into any, query string, args ...any) error
	SelectContext(ctx context.Context, into any, query string, args ...any) error
	Rebind(string) string
}

func selectx(ctx context.Context, db queryer, logger *log.Logger, debug bool, into any, query string, args ...any) error {
	start := time.Now()
	err := db.SelectContext(ctx, into, db.Rebind(query), args...)
	if debug {
		log.WithContext(ctx, logger).Debug("select done", "query.duration", time.Since(start), "query.raw", FormatQuery(query, args...))
	}
	return err
}

func get(ctx context.Context, db queryer, logger *log.Logger, debug bool, into any, query string, args ...any) error {
	start := time.Now()
	err := db.GetContext(ctx, into, db.Rebind(query), args...)
	if debug {
		log.WithContext(ctx, logger).Debug("get done", "query.duration", time.Since(start), "query.raw", FormatQuery(query, args...))
	}
	return err
}

func exec(ctx context.Context, db queryer, logger *log.Logger, debug bool, query string, args ...any) (sql.Result, error) {
	start := time.Now()
	res, err := db.ExecContext(ctx, db.Rebind(query), args...)
	if debug {
		log.WithContext(ctx, logger).Debug("exec done", "query.duration", time.Since(start), "query.raw", FormatQuery(query, args...))
	}
	return res, err
}
