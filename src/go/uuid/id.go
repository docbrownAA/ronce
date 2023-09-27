package uuid

import (
	"bytes"
	"crypto/rand"
	"database/sql/driver"
	"encoding/hex"
	"fmt"
	"ronce/src/go/errors"
	"strings"

	"github.com/oklog/ulid"
)

// ID is based on the oklog/ulid implementation. Those ids are not fully random
// generated. However, their textual representations do not match the db's uuid
// representation. To fix this, we override the Marshal and Driver interfaces
// to work with the compact uuid representation, without hyphens.
// The String, UnmarshalText and MarshalText methods are inspired from
// github.com/satori/go.uuid package.
// The textual reprensetations of those ids is "017f1bb5dd6ae6f7b3489d110fc6c286":
// - an hexa reprensentation of the id bytes
// - no hypen, which is also compatible with SQL queries
type ID [16]byte

func New() ID {
	return ID(ulid.MustNew(ulid.Now(), ulid.Monotonic(rand.Reader, 0)))
}

func Parse(raw string) (id ID, err error) {
	return id, id.UnmarshalText([]byte(raw))
}

func (id ID) String() string {
	return hex.EncodeToString([]byte(id[:]))
}

func (id *ID) UnmarshalText(raw []byte) error {
	src := bytes.Replace(raw[:], []byte{'-'}, nil, -1)
	dst := id[:]

	_, err := hex.Decode(dst, src)
	if err != nil {
		return fmt.Errorf("invalid id: expecting 16 bytes hex, got %w", err)
	}
	return nil
}

func (id ID) MarshalText() ([]byte, error) {
	dst := make([]byte, 32)
	hex.Encode(dst, id[:])
	return dst, nil
}

func (id ID) ULID() string {
	return ulid.ULID(id).String()
}

func (id *ID) Scan(in any) error {
	var err error
	switch t := in.(type) {
	case nil:
		return nil
	case []byte:
		err = id.UnmarshalText(bytes.ReplaceAll(t, []byte("-"), nil))
	case string:
		err = id.UnmarshalText([]byte(strings.ReplaceAll(t, "-", "")))
	default:
		err = errors.Newf("fail to scan %T => %T", in, id)
	}
	if err != nil {
		return errors.Newf("scan error: %w", err)
	}
	return nil
}

func (id ID) Value() (driver.Value, error) {
	return id.MarshalText()
}

func (id ID) IsZero() bool {
	return id == ID{}
}

func (id ID) ToGUID() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:])
}

type IDs []ID

func (s *IDs) UnmarshalText(raw []byte) error {
	for i, chunk := range bytes.Split(raw, []byte(`,`)) {
		if len(chunk) == 0 {
			continue
		}
		var id ID
		err := id.UnmarshalText(chunk)
		if err != nil {
			return fmt.Errorf(`parsing chunk %q at index %d: %w`, string(chunk), i, err)
		}
		*s = append(*s, id)
	}
	return nil
}

func (s IDs) MarshalText() ([]byte, error) {
	var buf = bytes.NewBuffer(nil)
	for i, id := range s {
		_, err := buf.WriteString(id.String())
		if err != nil {
			return nil, fmt.Errorf(`writing id %q at index %d: %w`, id, i, err)
		}
	}
	return buf.Bytes(), nil
}
