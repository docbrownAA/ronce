package sql

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"ronce/src/go/errors"
	"ronce/src/go/log"
	"ronce/src/go/timex"
	"ronce/src/go/uuid"
)

type Queue struct {
	Lanes     []string       `key:"lanes"     description:"lanes to consume"`
	Heartbeat timex.Duration `key:"heartbeat" description:"heartbeat interval of the consumer worker"`
	Timeout   timex.Duration `key:"timeout"   description:"timeout for running jobs to be claimable again"`
	Tries     int            `key:"tries"     description:"maximum number of retries for processing an interrupted job before marking it failed"`
}

type Status string

const (
	StatusPending    Status = "pending"
	StatusRunning    Status = "running"
	StatusSucceeded  Status = "succeeded"
	StatusFailed     Status = "failed"
	StatusCancelling Status = "cancelling"
	StatusCancelled  Status = "cancelled"
	StatusIgnored    Status = "ignored"
)

type Job struct {
	ID        uuid.ID   `db:"id"`
	Try       int       `db:"try"`
	Status    Status    `db:"status"`
	CreatedAt time.Time `db:"created_at"`
}

type JobRunner = func(context.Context, *log.Logger, json.RawMessage) bool

func (s Queue) Process(ctx context.Context, logger *log.Logger, db *DB, table string, run JobRunner) {
	t := time.NewTicker(time.Duration(s.Heartbeat))
	for {
		select {
		case <-ctx.Done():
			return
		case <-t.C:
		}

		var job struct {
			Job
			Payload json.RawMessage `db:"payload"` // payload is a jsonified SELECT * FROM table
		}
		now := time.Now()

		// Select the first pending job, or running that exceeded the timeout. We
		// have to use explicit locking in the subquery to avoid phantom reads. See
		// https://www.postgresql.org/docs/15/transaction-iso.html and
		// https://www.postgresql.org/docs/15/explicit-locking.html for details.
		query, args, err := In(fmt.Sprintf(`
			update %[1]s t
			set status = ?,
			    heartbeat_at = ?,
				try = try + 1
			where id in (
				select id
				from %[1]s
				where lane in (?)
				and (
					status = ?
					or (status = ? and heartbeat_at < ?)
				)
				order by array_position(array['%[2]s'], lane) asc, created_at asc
				limit 1
				for update
			)
			returning id, try, status, created_at, to_jsonb(t.*) as payload
		`, table, strings.Join(s.Lanes, "','")),
			StatusRunning,
			now,
			s.Lanes,
			StatusPending,
			StatusRunning,
			now.Add(-time.Duration(s.Timeout)),
		)
		if err != nil {
			logger.Error(`building job query`, `err`, err)
			continue
		}

		err = db.Get(ctx, &job, query, args...)
		if err == sql.ErrNoRows {
			continue
		}
		if err != nil {
			logger.Error(`retrieving job`, `err`, err)
			continue
		}

		logger := logger.With(`job_id`, job.ID)

		// We pre-increment the try counter in the job selection to avoid having to
		// do another query, so we decrement it here for checking the limit. If a
		// job has been interrupted more than the allowed number, mark it as failed.
		// This is mainly a failsafe for jobs that break the execution environment,
		// to avoid the retry mecanic to run wild.
		if job.Try-1 > s.Tries {
			job.Status = StatusFailed
			logger.Debug(`updating job status`, `status`, job.Status)
			_, err = db.Exec(ctx, fmt.Sprintf(`
			update %[1]s
			set status = ?
			where id = ?
		`, table), job.Status, job.ID)
			if err != nil {
				logger.Error(`updating job status`, `err`, err)
				continue
			}
		}

		job.Status = StatusRunning

		// Derive the app context for this job, and cancel it if we
		// detect that the job was cancelled. If the status changes for
		// something unexpected, we stop the analysis and propagate the
		// override.
		jobCtx, cancel := context.WithCancel(ctx)
		defer cancel()

		go func() {
			t := time.NewTicker(time.Duration(s.Heartbeat))
			for {
				select {
				case <-jobCtx.Done():
					logger.Debug(`closing monitoring routine`)
					return
				case <-t.C:
				}

				var status Status
				err = db.Get(ctx, &status, fmt.Sprintf(`
					update %[1]s
					set heartbeat_at = ?
					where id = ?
					returning status
				`, table), time.Now(), job.ID)

				// If the line can't be found anymore, this
				// means the line was removed. Cancel the
				// context and return to avoid leaks.
				if errors.Is(err, sql.ErrNoRows) {
					cancel()
					return
				}

				if err != nil {
					logger.Error(`monitoring job status`, `err`, err)
					continue
				}

				switch status {
				case StatusCancelling:
					logger.Debug(`cancellation detected`)
					cancel()
				case StatusRunning:
					continue
				default:
					job.Status = status
					cancel()
					continue
				}
			}
		}()

		ok := run(jobCtx, logger, job.Payload)
		switch {
		// If the status was manually overriden, log the event and skip
		// the job.
		case job.Status != StatusRunning:
			logger.Warn(`unexpected status detected`, `status`, job.Status)
			continue

		case ok:
			job.Status = StatusSucceeded

		// If we have an error and the parent context is closed, we
		// want to stop there and retry the job later, so we don't
		// change the status of the job.
		case ctx.Err() != nil:
			break

		// If we have an error and the job context is closed, the job
		// was cancelled by the user.
		case jobCtx.Err() != nil:
			job.Status = StatusCancelled

		default:
			job.Status = StatusFailed
		}

		logger.Debug(`updating job status`, `status`, job.Status)
		_, err = db.Exec(ctx, fmt.Sprintf(`
			update %[1]s
			set status = ?
			where id = ?
		`, table), job.Status, job.ID)
		if err != nil {
			logger.Error(`updating job status`, `err`, err)
			continue
		}
	}
}
