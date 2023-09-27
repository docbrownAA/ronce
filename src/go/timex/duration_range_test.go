package timex

import (
	"reflect"
	"testing"
)

func TestDurationRange_UnmarshalText(t *testing.T) {
	type Case struct {
		input string
		want  DurationRange
	}
	for _, c := range []Case{
		{`[0s,1s)`, DurationRange{0 * Second, 1 * Second}},
		{`[0s,4s)`, DurationRange{0 * Second, 4 * Second}},
		{`[1s,3s)`, DurationRange{1 * Second, 3 * Second}},
		{`[-1s,3s)`, DurationRange{-1 * Second, 3 * Second}},
		{`[-1s,-3s)`, DurationRange{-1 * Second, -3 * Second}},
	} {
		var got DurationRange
		if err := got.UnmarshalText([]byte(c.input)); err != nil {
			t.Error(err)
			t.Fail()
		}
		if !reflect.DeepEqual(c.want, got) {
			t.Errorf("UnmarshalText(%q): want %v, got %v", c.input, c.want, got)
			t.Fail()
		}
	}
}

func TestDurationRange_Intersection(t *testing.T) {
	type Case struct {
		d1, d2, d3 string
	}
	for _, c := range []Case{
		{`[0s,1s)`, `[0s,1s)`, `[0s,1s)`},
		{`[0s,0s)`, `[0s,1s)`, `[0s,0s)`},

		{`[0s,1s)`, `[2s,3s)`, `[0s,0s)`},
		{`[0s,4s)`, `[2s,3s)`, `[2s,3s)`},

		{`[0s,2s)`, `[1s,3s)`, `[1s,2s)`},
		{`[1s,3s)`, `[0s,2s)`, `[1s,2s)`},
	} {
		var d1, d2, d3 DurationRange
		if err := d1.UnmarshalText([]byte(c.d1)); err != nil {
			t.Error(err)
			t.Fail()
		}
		if err := d2.UnmarshalText([]byte(c.d2)); err != nil {
			t.Error(err)
			t.Fail()
		}
		if err := d3.UnmarshalText([]byte(c.d3)); err != nil {
			t.Error(err)
			t.Fail()
		}
		if d := d1.Intersection(d2); d != d3 {
			t.Errorf("intersection of %s and %s should be %s. got %s", d1, d2, d3, d)
			t.Fail()
		}
	}
}
