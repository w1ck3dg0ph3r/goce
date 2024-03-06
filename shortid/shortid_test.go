package shortid

import (
	"testing"
)

func TestShortid(t *testing.T) {
	ids := make(map[string]struct{}, 1000)
	for n := 0; n < 1000; n++ {
		id := New()
		if _, ok := ids[id.String()]; ok {
			t.Errorf("non-unique id: %q", id.String())
		}
		ids[id.String()] = struct{}{}
		parsed, err := Parse(id.String())
		if err != nil {
			t.Errorf("expected no error, got: %v", err)
		}
		if parsed != id {
			t.Errorf("parsed id not equal, expected: %v, got: %v", id, parsed)
		}
	}
}

func TestShortidParseInvalid(t *testing.T) {
	_, err := Parse("abcdef")
	if err != ErrInvalid {
		t.Errorf("expected ErrInvalid, got %v", err)
	}
}
