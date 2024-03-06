package shortid

import (
	"reflect"
	"testing"
)

func TestBase58(t *testing.T) {
	cases := []struct {
		in  []byte
		out string
	}{
		{nil, ""},
		{[]byte{}, ""},
		{[]byte("foo"), "WiWM"},
		{[]byte("quick brown fox jumped over lazy dog"), "NwsDvVT7Q9W3YGhmLwkeJhdZpF65RFkjSaYSQiQXnXegt4LXG"},
		{[]byte{0, 0, 0, 0, 0}, "YYYYY"},
		{[]byte{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, "ZAK2WAdzGUjM4"},
	}
	for _, tc := range cases {
		res := encode(tc.in)
		if res != tc.out {
			t.Errorf("encode(%v): expected: %q, got: %q", tc.in, tc.out, res)
		}
		back := decode(res)
		if len(back) == 0 && len(tc.in) == 0 {
			continue
		}
		if !reflect.DeepEqual(back, tc.in) {
			t.Errorf("decode(%q): expected: %v, got: %v", res, tc.in, back)
		}
	}
}
