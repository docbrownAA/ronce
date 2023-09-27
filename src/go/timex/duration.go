package timex

import (
	"database/sql/driver"
	"fmt"
	"time"
)

type Duration int64

const (
	Nanosecond  = Duration(time.Nanosecond)
	Microsecond = Duration(time.Microsecond)
	Millisecond = Duration(time.Millisecond)
	Second      = Duration(time.Second)
	Minute      = Duration(time.Minute)
	Hour        = Duration(time.Hour)
)

func Since(t Time) Duration {
	return Duration(time.Since(t))
}

func Until(t Time) Duration {
	return Duration(time.Until(t))
}

func (d Duration) Abs() Duration {
	return Duration(time.Duration(d).Abs())
}

func (d Duration) Hours() float64 {
	return time.Duration(d).Hours()
}

func (d Duration) Microseconds() int64 {
	return time.Duration(d).Microseconds()
}

func (d Duration) Milliseconds() int64 {
	return time.Duration(d).Milliseconds()
}

func (d Duration) Minutes() float64 {
	return time.Duration(d).Minutes()
}

func (d Duration) Nanoseconds() int64 {
	return time.Duration(d).Nanoseconds()
}

func (d Duration) Round(m Duration) Duration {
	return Duration(time.Duration(d).Round(time.Duration(m)))
}

func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

func (d Duration) String() string {
	return time.Duration(d).String()
}

func (d Duration) Truncate(m Duration) Duration {
	return Duration(time.Duration(d).Truncate(time.Duration(m)))
}

func (d *Duration) UnmarshalText(raw []byte) (err error) {
	dur, err := time.ParseDuration(string(raw))
	if err != nil {
		return err
	}
	*d = Duration(dur)

	return nil
}

func (d Duration) MarshalText() (res []byte, err error) {
	return []byte(time.Duration(d).String()), nil
}

func (d Duration) Value() (driver.Value, error) {
	return d.MarshalText()
}

func (d *Duration) Scan(src any) error {
	switch src := src.(type) {
	case string:
		return d.UnmarshalText([]byte(src))
	case []byte:
		return d.UnmarshalText(src)
	default:
		return fmt.Errorf(`invalid scan pair %T => %T`, src, d)
	}
}

func (d *Duration) SubtitleFormat() string {
	ms := (d.Milliseconds() - int64(d.Seconds())*1000)
	v := fmt.Sprintf("%d:%d:%d,%d", int64(d.Hours()), int64(d.Minutes()), int64(d.Seconds()), ms)
	return v
}
