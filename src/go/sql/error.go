package sql

import (
	"database/sql"
	"fmt"
	"ronce/src/go/errors"
	"strings"

	"github.com/lib/pq"
)

var ErrNoRows = sql.ErrNoRows

// List of error codes, from https://www.postgresql.org/docs/8.2/errcodes-appendix.html
const (
	CodeDuplicateKeyValue   = "23505"
	CodeForeignKeyViolation = "23503"
	CodeNoNullViolation     = "23502"
)

// ParseForeignKeyViolationDetail returns the value responsible for the KeyViolation.
func ParseForeignKeyViolationDetail(pqErr *pq.Error) (key, value string, err error) {
	if !(pqErr.Code == CodeForeignKeyViolation || pqErr.Code == CodeDuplicateKeyValue) {
		return "", "", fmt.Errorf("call to GetErrorConstraint on error with code %q", pqErr.Code)
	}

	raw := strings.TrimPrefix(pqErr.Detail, "Key (")
	chunks := strings.Split(raw, ")=(")
	key = chunks[0]
	if len(chunks) != 2 {
		return "", "", errors.Newf("cannot parse err.Details: unexpected format %q", pqErr.Detail)
	}
	switch pqErr.Code {
	// "Key (user_id)=(017ec024-ff47-f306-2467-b1a68798b880) is not present in table \"user\"."
	case CodeForeignKeyViolation:
		chunks = strings.Split(chunks[1], ") is not")
		value = chunks[0]

	// "Key (name)=(Zoubidou) already exist"
	case CodeDuplicateKeyValue:
		chunks = strings.Split(chunks[1], ") already")
		value = chunks[0]
	}

	return key, value, nil
}
