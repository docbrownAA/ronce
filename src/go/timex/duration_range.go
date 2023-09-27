package timex

import (
	"bytes"
	"fmt"
	"ronce/src/go/errors"
)

type DurationRange struct {
	Start Duration
	End   Duration
}

func (d DurationRange) Validate() error {
	if d.Start > d.End {
		return errors.New("duration_range has start after its end")
	}
	return nil
}

func (r DurationRange) String() string {
	raw, _ := r.MarshalText()
	return string(raw)
}

func (r DurationRange) Duration() Duration {
	return r.End - r.Start
}

func (r DurationRange) MarshalText() ([]byte, error) {
	return []byte(fmt.Sprintf(`[%s,%s)`, r.Start.String(), r.End.String())), nil
}

func (r *DurationRange) UnmarshalText(src []byte) error {
	if src[0] != '[' {
		return errors.Newf(`invalid range: expected starting [, found %c`, src[0])
	}
	if src[len(src)-1] != ')' {
		return errors.Newf(`invalid range: expected ending (, found %c`, src[len(src)-1])
	}
	src = src[1 : len(src)-1]

	chunks := bytes.Split(src, []byte(","))
	if len(chunks) != 2 {
		return errors.Newf(`invalid range: expected 2 segments, got %d`, len(chunks))
	}

	err := r.Start.UnmarshalText(chunks[0])
	if err != nil {
		return errors.Wrap(err, `parsing range start`)
	}

	err = r.End.UnmarshalText(chunks[1])
	if err != nil {
		return errors.Wrap(err, `parsing range end`)
	}

	return nil
}

// Intersection of two ranges.
func (d DurationRange) Intersection(d2 DurationRange) DurationRange {
	var out DurationRange = d
	if d2.Start > out.Start {
		out.Start = d2.Start
	}
	if d2.End < out.End {
		out.End = d2.End
	}
	if out.End < out.Start {
		out = DurationRange{}
	}
	return out
}
